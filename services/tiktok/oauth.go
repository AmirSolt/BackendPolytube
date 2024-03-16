package tiktok

import (
	"basedpocket/base"
	"basedpocket/cmodels"
	"basedpocket/utils"
	"fmt"
	"net/http"
	"net/url"

	"github.com/carlmjohnson/requests"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/tools/types"
)

// ====================================
// ====================================
// ====================================

func fetchAndStoreAccessToken(app core.App, ctx echo.Context, env *base.Env, code string) *utils.CError {
	// ========================
	// fetch new access token
	formData := url.Values{}
	formData.Add("client_key", env.TIKTOK_CLIENT_KEY)
	formData.Add("client_secret", env.TIKTOK_CLIENT_SECRET)
	formData.Add("code", code)
	formData.Add("grant_type", "authorization_code")
	formData.Add("redirect_uri", fmt.Sprintf("%s/platforms/tiktok/oauth-success", env.DOMAIN))
	// handle response
	resRaw := &TikTokAccessTokenResponseRaw{}
	err := requests.
		URL("https://open.tiktokapis.com/v2/oauth/token/").
		Method(http.MethodPost).
		BodyForm(formData).
		ToJSON(&resRaw).
		Fetch(ctx.Request().Context())
	if err != nil {
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	// convert raw response
	var res *TikTokAccessTokenResponse
	if err := convertTiktokAccessTokenResponse(resRaw, res); err != nil {
		return nil
	}
	// =================
	// upsert db
	appError := upsertTiktokDBOnNewAccess(app, ctx, env, res)
	if appError != nil {
		return appError
	}

	return nil
}

// ====================================
func refreshAccessToken(app core.App, ctx echo.Context, env *base.Env, channel *cmodels.Channel) *utils.CError {
	// ==========================
	// get user
	var user *cmodels.User
	if err := user.GetUserByContext(ctx); err != nil {
		return err
	}

	var oauth *cmodels.OAuth
	if err := oauth.FindOAuth(app, &cmodels.FindOAuthParams{User: user.Id, Channel: oauth.Id}); err != nil {
		return err
	}

	formData := url.Values{}
	formData.Add("client_key", env.TIKTOK_CLIENT_KEY)
	formData.Add("client_secret", env.TIKTOK_CLIENT_SECRET)
	formData.Add("grant_type", "refresh_token")
	formData.Add("refresh_token", oauth.RefreshToken)

	resRaw := &TikTokAccessTokenResponseRaw{}
	err := requests.
		URL("https://open.tiktokapis.com/v2/oauth/token/").
		Method(http.MethodPost).
		BodyForm(formData).
		ToJSON(&resRaw).
		Fetch(ctx.Request().Context())
	if err != nil {
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID}
	}
	// convert raw response
	var res *TikTokAccessTokenResponse
	if err := convertTiktokAccessTokenResponse(resRaw, res); err != nil {
		return nil
	}
	// ===============
	appError := upsertTiktokDBOnNewAccess(app, ctx, env, res)
	if appError != nil {
		return appError
	}

	return nil
}

// ============================================

func upsertTiktokDBOnNewAccess(app core.App, ctx echo.Context, env *base.Env, response *TikTokAccessTokenResponse) *utils.CError {
	// ==========================
	// get user
	var user *cmodels.User
	if err := user.GetUserByContext(ctx); err != nil {
		return err
	}

	// ==========================
	// find channel
	var channel *cmodels.Channel
	if err := channel.Find(app, &cmodels.FindChannelParams{User: user.Id, ExternalID: response.OpenID}); err != nil {
		return err
	}
	var oauth *cmodels.OAuth
	if err := oauth.FindOAuth(app, &cmodels.FindOAuthParams{User: user.Id, Channel: oauth.Id}); err != nil {
		return err
	}
	// ==========================
	// start transaction
	err := app.Dao().RunInTransaction(func(txDao *daos.Dao) error {

		if channel != nil {
			// ==========================
			// update channel
			channel.AccessExpiresIn = response.AccessTokenExpiresIn
			if appError := channel.Save(app); appError != nil {
				return fmt.Errorf("points to eventID: %s", appError.EventID)
			}
			// ==========================
			// update oauth
			oauth.Scope = response.Scope
			oauth.AccessToken = response.AccessToken
			oauth.AccessTokenExpiresIn = response.AccessTokenExpiresIn
			oauth.RefreshToken = response.RefreshToken
			oauth.RefreshTokenExpiresIn = response.RefreshTokenExpiresIn
			if appError := oauth.SaveOAuth(app); appError != nil {
				return fmt.Errorf("points to eventID: %s", appError.EventID)
			}
		} else {
			// ==========================
			// new channel
			newChannel := &cmodels.Channel{
				User:            user.Id,
				PlatformName:    string(cmodels.TikTokPlatform),
				ExternalID:      response.OpenID,
				AccessExpiresIn: response.AccessTokenExpiresIn,
			}
			if appError := newChannel.Save(app); appError != nil {
				return fmt.Errorf("points to eventID: %s", appError.EventID)
			}
			// ==========================
			// new oauth
			newOAuth := &cmodels.OAuth{
				User:                  user.Id,
				Channel:               newChannel.Id,
				Scope:                 response.Scope,
				AccessToken:           response.AccessToken,
				AccessTokenExpiresIn:  response.AccessTokenExpiresIn,
				RefreshToken:          response.RefreshToken,
				RefreshTokenExpiresIn: response.RefreshTokenExpiresIn,
			}
			if appError := newOAuth.SaveOAuth(app); appError != nil {
				return fmt.Errorf("points to eventID: %s", appError.EventID)
			}
		}

		return nil
	})

	if err != nil {
		// insert failed activity
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID}
	}

	return nil
}

// ====================================
// ====================================
// ====================================
type TikTokAccessTokenResponseRaw struct {
	OpenID                string `json:"open_id"`
	Scope                 string `json:"scope"`
	AccessToken           string `json:"access_token"`
	AccessTokenExpiresIn  int64  `json:"expires_in"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn int64  `json:"refresh_expires_in"`
	TokenType             string `json:"token_type"`
}

type TikTokAccessTokenResponse struct {
	OpenID                string          `json:"open_id"`
	Scope                 string          `json:"scope"`
	AccessToken           string          `json:"access_token"`
	AccessTokenExpiresIn  *types.DateTime `json:"expires_in"`
	RefreshToken          string          `json:"refresh_token"`
	RefreshTokenExpiresIn *types.DateTime `json:"refresh_expires_in"`
	TokenType             string          `json:"token_type"`
}

func convertTiktokAccessTokenResponse(raw *TikTokAccessTokenResponseRaw, res *TikTokAccessTokenResponse) *utils.CError {
	accessTokenExpiresIn, err := types.ParseDateTime(raw.AccessTokenExpiresIn)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	refreshTokenExpiresIn, err2 := types.ParseDateTime(raw.RefreshTokenExpiresIn)
	if err2 != nil {
		eventID := sentry.CaptureException(err2)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err2}
	}
	res = &TikTokAccessTokenResponse{
		OpenID:                raw.OpenID,
		Scope:                 raw.Scope,
		AccessToken:           raw.AccessToken,
		AccessTokenExpiresIn:  &accessTokenExpiresIn,
		RefreshToken:          raw.RefreshToken,
		RefreshTokenExpiresIn: &refreshTokenExpiresIn,
		TokenType:             raw.TokenType,
	}
	return nil
}
