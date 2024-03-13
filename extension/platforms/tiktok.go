package platforms

import (
	"basedpocket/extension"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"

	"github.com/spf13/cast"
)

// ====================================
// ====================================
// ====================================

func handleOAuthRequest(app core.App, ctx echo.Context, env *extension.Env) error {

	csrfState, err := generateCSRFState()
	if err != nil {
		eventID := sentry.CaptureException(err)
		return ctx.JSON(http.StatusInternalServerError, extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)})
	}
	ctx.SetCookie(&http.Cookie{
		Name:   "csrfState",
		Value:  csrfState,
		MaxAge: 60,
	})

	queries := map[string]string{
		"client_key":    env.TIKTOK_CLIENT_KEY,
		"scope":         "user.info.basic,user.info.profile,user.info.stats,video.list,video.publish,video.upload",
		"response_type": "code",
		"redirect_uri":  fmt.Sprintf("%s/platforms/tiktok/oauth-success", env.DOMAIN),
		"state":         csrfState,
	}

	url, err := buildURLFromMap("https://www.tiktok.com/v2/auth/authorize?", queries)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return ctx.JSON(http.StatusInternalServerError, extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)})
	}

	return ctx.Redirect(http.StatusTemporaryRedirect, url)
}

// ====================================
// ====================================
// ====================================

type TikTokAuthorizationResponse struct {
	Code             string `json:"code"`
	Scopes           string `json:"scopes"`
	State            string `json:"state"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func handleOAuthSuccess(app core.App, ctx echo.Context, env *extension.Env) error {

	resp := new(TikTokAuthorizationResponse)
	if err := ctx.Bind(resp); err != nil {
		eventID := sentry.CaptureException(err)
		return ctx.JSON(http.StatusInternalServerError, extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)})
	}

	if resp.Error != "" {
		err := fmt.Errorf("error: %s | %s", resp.Error, resp.ErrorDescription)
		eventID := sentry.CaptureException(err)
		return ctx.JSON(http.StatusInternalServerError, extension.AppError{Message: err.Error(), EventID: fmt.Sprintf("%v", &eventID)})
	}

	go fetchAndStoreAccessToken(app, ctx, env, resp.Code)
	return ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/user", env.FRONTEND_DOMAIN))
}

// ====================================
// ====================================
// ====================================

func handleRefreshToken(app core.App, env *extension.Env) {
	// oath2.revokeAccessToken()
}

// ====================================
// ====================================
// ====================================

func handleRevokeToken(app core.App, ctx echo.Context, env *extension.Env) error {
	// oath2.revokeAccessToken()
	return ctx.String(http.StatusOK, "Hello, World!")
}

// ====================================
// ====================================
// ====================================

type TikTokAccessTokenResponse struct {
	OpenID           string `json:"open_id"`
	Scope            string `json:"scope"`
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
}

func fetchAndStoreAccessToken(app core.App, ctx echo.Context, env *extension.Env, code string) *extension.AppError {

	formData := map[string]string{
		"client_key":    env.TIKTOK_CLIENT_KEY,
		"client_secret": env.TIKTOK_CLIENT_SECRET,
		"code":          code,
		"grant_type":    "authorization_code",
		"redirect_uri":  fmt.Sprintf("%s/platforms/tiktok/oauth-success", env.DOMAIN),
	}

	// Encode request body as url-encoded form
	data, err := json.Marshal(formData)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}

	// POST request to TikTok API
	req, err := http.NewRequest(http.MethodPost, "https://open.tiktokapis.com/v2/oauth/token/", bytes.NewBuffer(data))
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationForm)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}

	client := &http.Client{}
	resRaw, err := client.Do(req)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}
	defer resRaw.Body.Close()

	if resRaw.StatusCode != http.StatusCreated {
		eventID := sentry.CaptureException(err)
		return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}
	res := &TikTokAccessTokenResponse{}
	if err := json.NewDecoder(resRaw.Body).Decode(res); err != nil {
		eventID := sentry.CaptureException(err)
		return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}

	appError := upsertTiktokOAuth(app, ctx, env, res)
	if appError != nil {
		return appError
	}

	return nil
}

func upsertTiktokOAuth(app core.App, ctx echo.Context, env *extension.Env, response *TikTokAccessTokenResponse) *extension.AppError {
	user, _ := ctx.Get(apis.ContextAuthRecordKey).(*models.Record)

	// ==========================
	// find it
	tiktokOauth, err := app.Dao().FindFirstRecordByData("tiktok_oauths", "user", user.Get("id"))
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		eventID := sentry.CaptureException(err)
		return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}

	if tiktokOauth != nil {
		// ==========================
		// update
		tiktokOauth.Set("open_id", response.OpenID)
		tiktokOauth.Set("scope", response.Scope)
		tiktokOauth.Set("access_token", response.AccessToken)
		tiktokOauth.Set("access_token_expires_in", cast.ToTime(response.ExpiresIn))
		tiktokOauth.Set("refresh_token", response.RefreshToken)
		tiktokOauth.Set("refresh_token_expires_in", cast.ToTime(response.RefreshExpiresIn))
		if err := app.Dao().SaveRecord(tiktokOauth); err != nil {
			eventID := sentry.CaptureException(err)
			return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
		}
	} else {
		// ==========================
		// insert
		tiktokOauthCollection, err := app.Dao().FindCollectionByNameOrId("tiktok_oauths")
		if err != nil {
			eventID := sentry.CaptureException(err)
			return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
		}

		newTiktokOauth := models.NewRecord(tiktokOauthCollection)

		newTiktokOauth.Set("user", user.Id)
		newTiktokOauth.Set("open_id", response.OpenID)
		newTiktokOauth.Set("scope", response.Scope)
		newTiktokOauth.Set("access_token", response.AccessToken)
		newTiktokOauth.Set("access_token_expires_in", cast.ToTime(response.ExpiresIn))
		newTiktokOauth.Set("refresh_token", response.RefreshToken)
		newTiktokOauth.Set("refresh_token_expires_in", cast.ToTime(response.RefreshExpiresIn))
		if err := app.Dao().SaveRecord(newTiktokOauth); err != nil {
			eventID := sentry.CaptureException(err)
			return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
		}
	}

	return nil
}

// ====================================
// ====================================
// ====================================

func refreshAccessToken(app core.App, ctx echo.Context, env *extension.Env) error {

	return ctx.String(http.StatusOK, "Hello, World!")
}

// ====================================
// ====================================
// ====================================

func revokeAccessToken(app core.App, ctx echo.Context, env *extension.Env) error {

	return ctx.String(http.StatusOK, "Hello, World!")
}
