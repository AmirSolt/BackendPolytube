package platforms

import (
	"basedpocket/extension"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/models"

	"github.com/spf13/cast"
)

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
func handleRevokeToken(app core.App, ctx echo.Context, env *extension.Env) error {
	platform_account_id := ctx.PathParam("platform_account_id")
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

	appError := upsertTiktokAccountInfo(app, ctx, env, res)
	if appError != nil {
		return appError
	}

	return nil
}
func upsertTiktokAccountInfo(app core.App, ctx echo.Context, env *extension.Env, response *TikTokAccessTokenResponse) *extension.AppError {
	// ==========================
	// get user
	user := ctx.Get(apis.ContextAuthRecordKey).(*models.Record)
	if user == nil {
		err := fmt.Errorf("user not found")
		eventID := sentry.CaptureException(err)
		return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}

	// ==========================
	// get platform account
	account, err := app.Dao().FindFirstRecordByFilter("platform_accounts",
		fmt.Sprintf("external_account_id = '%s'", response.OpenID),
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		eventID := sentry.CaptureException(err)
		return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}

	if account != nil {
		// ==========================
		// start transaction
		err := app.Dao().RunInTransaction(func(txDao *daos.Dao) error {

			// ========================
			// if account access is not expired, make activity status warning
			isAccountAccessExpired := time.Now().After(account.Get("access_expires_in").(time.Time))
			activityStatus := SuccessStatus
			if !isAccountAccessExpired {
				activityStatus = WarningStatus
			}

			// ==========================
			// update platform account
			appError := updatePlatformAccount(app, account, response)
			if appError != nil {
				return fmt.Errorf("points to eventID: %s", appError.EventID)
			}
			appErr := InsertPlatformActivity(app, ctx, AcitvityParams{
				PlatformAccountID: account.Id,
				Title:             "Tiktok Internal Account Updated",
				Message:           "N/A",
				Status:            activityStatus,
			})
			if appErr != nil {
				return fmt.Errorf("points to eventID: %s", appErr.EventID)
			}
			// ==========================
			// update platform account
			oauthError := updateOAuth(app, account, response)
			if oauthError != nil {
				return fmt.Errorf("points to eventID: %s", oauthError.EventID)
			}
			accErr := InsertPlatformActivity(app, ctx, AcitvityParams{
				PlatformAccountID: account.Id,
				Title:             "Tiktok Permission Updated",
				Message:           "N/A",
				Status:            activityStatus,
			})
			if accErr != nil {
				return fmt.Errorf("points to eventID: %s", appErr.EventID)
			}

			return nil
		})

		if err != nil {
			eventID := sentry.CaptureException(err)
			return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
		}

	} else {
		// ==========================
		// start transaction
		err := app.Dao().RunInTransaction(func(txDao *daos.Dao) error {

			// ==========================
			// insert platform account
			newAccount, appError := insertPlatformAccount(app, user, response)
			if appError != nil {
				return fmt.Errorf("points to eventID: %v", appError.EventID)
			}
			appErr := InsertPlatformActivity(app, ctx, AcitvityParams{
				PlatformAccountID: newAccount.Id,
				Title:             "Tiktok Internal Account Created",
				Message:           "N/A",
				Status:            SuccessStatus,
			})
			if appErr != nil {
				return fmt.Errorf("points to eventID: %v", appErr.EventID)
			}
			// ==========================
			// insert platform account
			oauthError := insertOAuth(app, user, newAccount, response)
			if oauthError != nil {
				return fmt.Errorf("points to eventID: %v", oauthError.EventID)
			}
			accErr := InsertPlatformActivity(app, ctx, AcitvityParams{
				PlatformAccountID: newAccount.Id,
				Title:             "Tiktok Permission Created",
				Message:           "N/A",
				Status:            SuccessStatus,
			})
			if accErr != nil {
				return fmt.Errorf("points to eventID: %v", appErr.EventID)
			}

			return nil
		})

		if err != nil {
			eventID := sentry.CaptureException(err)
			return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
		}

	}

	return nil
}

// ====================================
func refreshAccessToken(app core.App, ctx echo.Context, env *extension.Env) error {

	return ctx.String(http.StatusOK, "Hello, World!")
}

// ====================================
func revokeAccessToken(app core.App, ctx echo.Context, env *extension.Env) error {

	return ctx.String(http.StatusOK, "Hello, World!")
}

// ====================================
// ====================================
// ====================================
func insertPlatformAccount(app core.App, user *models.Record, response *TikTokAccessTokenResponse) (*models.Record, *extension.AppError) {
	collection, err := app.Dao().FindCollectionByNameOrId("platform_accounts")
	if err != nil {
		eventID := sentry.CaptureException(err)
		return nil, &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}

	record := models.NewRecord(collection)
	record.Set("user", user.Id)
	record.Set("platform_name", TikTokPlatform)
	record.Set("external_account_id", response.OpenID)
	record.Set("access_can_expire", true)
	record.Set("access_expires_in", cast.ToTime(response.RefreshExpiresIn))
	if err := app.Dao().SaveRecord(record); err != nil {
		eventID := sentry.CaptureException(err)
		return nil, &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}
	return record, nil
}
func updatePlatformAccount(app core.App, account *models.Record, response *TikTokAccessTokenResponse) *extension.AppError {
	record, err := app.Dao().FindFirstRecordByFilter("platform_accounts",
		fmt.Sprintf("id = '%s'", account.Id),
	)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}
	record.Set("access_expires_in", cast.ToTime(response.RefreshExpiresIn))
	if err := app.Dao().SaveRecord(record); err != nil {
		eventID := sentry.CaptureException(err)
		return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}
	return nil
}
func insertOAuth(app core.App, user *models.Record, account *models.Record, response *TikTokAccessTokenResponse) *extension.AppError {
	tiktokOauthCollection, err := app.Dao().FindCollectionByNameOrId("oauths")
	if err != nil {
		eventID := sentry.CaptureException(err)
		return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}

	newTiktokOauth := models.NewRecord(tiktokOauthCollection)

	newTiktokOauth.Set("user", user.Id)
	newTiktokOauth.Set("account_id", response.OpenID)
	newTiktokOauth.Set("scope", response.Scope)
	newTiktokOauth.Set("access_token", response.AccessToken)
	newTiktokOauth.Set("access_token_expires_in", cast.ToTime(response.ExpiresIn))
	newTiktokOauth.Set("refresh_token", response.RefreshToken)
	newTiktokOauth.Set("refresh_token_expires_in", cast.ToTime(response.RefreshExpiresIn))
	if err := app.Dao().SaveRecord(newTiktokOauth); err != nil {
		eventID := sentry.CaptureException(err)
		return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}
	return nil
}
func updateOAuth(app core.App, account *models.Record, response *TikTokAccessTokenResponse) *extension.AppError {
	tiktokOauth, err := app.Dao().FindFirstRecordByFilter("oauths",
		fmt.Sprintf("platform_account = '%s'", account.Id),
	)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}
	tiktokOauth.Set("scope", response.Scope)
	tiktokOauth.Set("access_token", response.AccessToken)
	tiktokOauth.Set("access_token_expires_in", cast.ToTime(response.ExpiresIn))
	tiktokOauth.Set("refresh_token", response.RefreshToken)
	tiktokOauth.Set("refresh_token_expires_in", cast.ToTime(response.RefreshExpiresIn))
	if err := app.Dao().SaveRecord(tiktokOauth); err != nil {
		eventID := sentry.CaptureException(err)
		return &extension.AppError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}
	return nil
}
