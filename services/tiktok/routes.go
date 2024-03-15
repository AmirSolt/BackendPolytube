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
)

func handleOAuthRequest(app core.App, ctx echo.Context, env *base.Env) error {

	csrfState, err := utils.GenerateCSRFState()
	if err != nil {
		eventID := sentry.CaptureException(err)
		return ctx.JSON(http.StatusInternalServerError, utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err})
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

	url, err := utils.BuildURLFromMap("https://www.tiktok.com/v2/auth/authorize?", queries)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return ctx.JSON(http.StatusInternalServerError, utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err})
	}

	return ctx.Redirect(http.StatusTemporaryRedirect, url)
}

// ====================================

func handleOAuthSuccess(app core.App, ctx echo.Context, env *base.Env) error {

	// handle response
	resp := new(TikTokAuthorizationResponseRaw)
	if err := ctx.Bind(resp); err != nil {
		eventID := sentry.CaptureException(err)
		return ctx.JSON(http.StatusInternalServerError, utils.CError{Message: "Internal Server Error", EventID: *eventID})
	}
	if resp.Error != "" {
		err := fmt.Errorf("error: %s | %s", resp.Error, resp.ErrorDescription)
		eventID := sentry.CaptureException(err)
		return ctx.JSON(http.StatusInternalServerError, utils.CError{Message: err.Error(), EventID: *eventID, Error: err})
	}

	// fetch and store access token
	go fetchAndStoreAccessToken(app, ctx, env, resp.Code)

	return ctx.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s/user", env.FRONTEND_DOMAIN))
}

// ====================================
func handleRevokeToken(app core.App, ctx echo.Context, env *base.Env) error {

	// ==========================
	// get user
	var user *cmodels.User
	if err := user.GetUserByContext(ctx); err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	channelID := ctx.PathParam("channel_id")
	if channelID == "" {
		err := fmt.Errorf("channel_id is empty")
		eventID := sentry.CaptureException(err)
		return ctx.JSON(http.StatusInternalServerError, utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err})
	}
	var oauth *cmodels.OAuth
	if err := oauth.FindOAuth(app, &cmodels.FindOAuthParams{User: user.Id, Channel: channelID}); err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	// ===================
	// request revoke
	formData := url.Values{}
	formData.Add("client_key", env.TIKTOK_CLIENT_KEY)
	formData.Add("client_secret", env.TIKTOK_CLIENT_SECRET)
	formData.Add("token", oauth.AccessToken)

	errReq := requests.
		URL("https://open.tiktokapis.com/v2/oauth/revoke/").
		Method(http.MethodPost).
		BodyForm(formData).
		CheckStatus(http.StatusOK, http.StatusAccepted).
		Fetch(ctx.Request().Context())
	if errReq != nil {
		eventID := sentry.CaptureException(errReq)
		return ctx.JSON(http.StatusInternalServerError, utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: errReq})
	}

	return ctx.NoContent(http.StatusOK)
}

// ====================================
// ====================================
// ====================================

type TikTokAuthorizationResponseRaw struct {
	Code             string `json:"code"`
	Scopes           string `json:"scopes"`
	State            string `json:"state"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}
