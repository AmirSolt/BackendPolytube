package elevenlabs

import (
	"basedpocket/base"
	"basedpocket/cmodels"
	"basedpocket/utils"
	"net/http"
	"net/url"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

type DubbingResponse struct {
	DubbingID           string `json:"dubbing_id"`
	ExpectedDurationSec int    `json:"expected_duration_sec"`
}

func GetDubFileURL() {
	"https://api.elevenlabs.io/v1/dubbing/{dubbing_id}/audio/{language_code}"
}

func RequestAndUpdateDubjob(app core.App, ctx echo.Context, env *base.Env, dubjob *cmodels.Dubjob) *utils.CError {

	formData := url.Values{}
	formData.Add("mode", "automatic")
	formData.Add("source_url", dubjob.SourceURL)
	formData.Add("target_lang", dubjob.TargetLanguage)

	// handle response
	res := &DubbingResponse{}
	err := requests.
		URL("https://api.elevenlabs.io/v1/dubbing/").
		ContentType("multipart/form-data").
		Header("xi-api-key", env.ELEVENLABS_API_KEY).
		Method(http.MethodPost).
		BodyForm(formData).
		ToJSON(&res).
		Fetch(ctx.Request().Context())
	if err != nil {
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}

	// ===============
	// convert time to datetime
	expectedIn, err := types.ParseDateTime(time.Now().Add(time.Second * time.Duration(res.ExpectedDurationSec)))
	if err != nil {
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	dubjob.ExternalID = res.DubbingID
	dubjob.ExpectedReadyIn = &expectedIn
	if err := dubjob.SaveDubjob(app); err != nil {
		return err
	}

	return nil
}
