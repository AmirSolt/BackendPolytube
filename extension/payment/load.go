package payment

import (
	"basedpocket/extension"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func LoadPayment(app *pocketbase.PocketBase, env *extension.Env) {

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		createCustomersCollection(e.App)
		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/webhooks/stripe",
			Handler: func(c echo.Context) error {
				return handleStripeWebhook(e.App, c, env)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(e.App),
			},
		})

		return nil
	})
}