package platforms

import (
	"basedpocket/extension"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/core"
)

// ====================================
// ====================================
// ====================================

func handleOAuthRequest(app core.App, ctx echo.Context, env *extension.Env) error {

	csrfState, err := generateCSRFState()
	if err != nil {
		eventId := sentry.CaptureException(err)
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("Failed to generate CSRF state. EventID: %v", &eventId))
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
		eventId := sentry.CaptureException(err)
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("Failed to build url. EventID: %v", &eventId))
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
		eventId := sentry.CaptureException(err)
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("Handling form data failed. EventID: %v", &eventId))
	}

	if resp.Error != "" {
		err := fmt.Errorf("error: %s | %s", resp.Error, resp.ErrorDescription)
		eventId := sentry.CaptureException(err)
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("%s. EventID: %v", err.Error(), &eventId))
	}

	// async oath2.fetchAccessToken()
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

func fetchAccessToken(app core.App, ctx echo.Context, env *extension.Env, code string) *sentry.EventID {

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
		return sentry.CaptureException(err)
	}

	// POST request to TikTok API
	req, err := http.NewRequest(http.MethodPost, "https://open.tiktokapis.com/v2/oauth/token/", bytes.NewBuffer(data))
	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationForm)
	if err != nil {
		return sentry.CaptureException(err)
	}

	client := &http.Client{}
	resRaw, err := client.Do(req)
	if err != nil {
		return sentry.CaptureException(err)
	}
	defer resRaw.Body.Close()

	// Read and process response body (replace with your logic)
	res := &TikTokAccessTokenResponse{}
	if err := json.NewDecoder(resRaw.Body).Decode(res); err != nil {
		return sentry.CaptureException(err)
	}
	if resRaw.StatusCode != http.StatusCreated {
		return sentry.CaptureException(err)
	}

	// save token to db
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
