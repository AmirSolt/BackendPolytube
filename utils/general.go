package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"
	"reflect"

	"github.com/getsentry/sentry-go"
)

func GetFieldTag(obj interface{}, fieldName string, tagName string) (string, *CError) {
	field, ok := reflect.TypeOf(obj).Elem().FieldByName(fieldName)
	if !ok {
		err := fmt.Errorf("struct reflect failed. Struct: %+v | fieldName: %s | tagName: %s", obj, fieldName, tagName)
		eventID := sentry.CaptureException(err)
		return "", &CError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}
	return string(field.Tag.Get(tagName)), nil
}

// ==========================

type PlatformName string

const TikTokPlatform PlatformName = "tiktok"
const YoutubePlatform PlatformName = "youtube"

// ==========================

func buildURLFromMap(baseURL string, queries map[string]string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	for key, value := range queries {
		q.Add(key, value)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func generateCSRFState() (string, error) {
	// Create a byte slice to hold the random bytes
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Encode the random bytes to hex string (base 16) and remove the "0x" prefix
	return hex.EncodeToString(b)[2:], nil
}

// ==========================

type AcitvityParams struct {
	PlatformAccountID string
	ExternalAccountID string
	Title             string
	Message           string
	Status            ActivityStatus
}

type ActivityStatus string

const PrimaryStatus ActivityStatus = "primary"
const SecondaryStatus ActivityStatus = "secondary"
const TertiaryStatus ActivityStatus = "tertiary"
const SuccessStatus ActivityStatus = "success"
const WarningStatus ActivityStatus = "warning"
const ErrorStatus ActivityStatus = "error"
const SurfaceStatus ActivityStatus = "surface"

// func InsertPlatformActivity(app core.App, ctx echo.Context, actParams AcitvityParams) *utils.CError {
// 	// ============================
// 	// get user
// 	user := ctx.Get(apis.ContextAuthRecordKey).(*models.Record)
// 	if user == nil {
// 		err := fmt.Errorf("user not found")
// 		eventID := sentry.CaptureException(err)
// 		return &utils.CError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
// 	}

// 	// ============================
// 	// get platform account id
// 	var platformAccountID string = ""
// 	if actParams.PlatformAccountID == "" {
// 		account, err := app.Dao().FindFirstRecordByFilter("channels",
// 			fmt.Sprintf("external_account_id = '%s'", actParams.ExternalAccountID),
// 		)
// 		if err != nil && !errors.Is(err, sql.ErrNoRows) {
// 			eventID := sentry.CaptureException(err)
// 			return &utils.CError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
// 		}
// 		platformAccountID = account.Id
// 	} else {
// 		platformAccountID = actParams.PlatformAccountID
// 	}

// 	// ============================
// 	// start insert
// 	collection, err := app.Dao().FindCollectionByNameOrId("platform_activities")
// 	if err != nil {
// 		eventID := sentry.CaptureException(err)
// 		return &utils.CError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
// 	}

// 	record := models.NewRecord(collection)

// 	record.Set("user", user.Id)
// 	record.Set("platform_account", platformAccountID)
// 	record.Set("title", actParams.Message)
// 	record.Set("message", actParams.Title)
// 	record.Set("status", actParams.Status)
// 	if err := app.Dao().SaveRecord(record); err != nil {
// 		eventID := sentry.CaptureException(err)
// 		return &utils.CError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
// 	}

// 	return nil
// }
