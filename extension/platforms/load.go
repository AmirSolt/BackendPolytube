package platforms

import (
	"basedpocket/extension"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/cron"
)

func LoadPlatforms(app *pocketbase.PocketBase, env *extension.Env) {
	loadTiktok(app, env)
}

func loadTiktok(app *pocketbase.PocketBase, env *extension.Env) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/platforms/tiktok/oauth-request",
			Handler: func(c echo.Context) error {
				return handleOAuthRequest(e.App, c, env)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(e.App),
			},
		})

		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/platforms/tiktok/oauth-success",
			Handler: func(c echo.Context) error {
				return handleOAuthSuccess(e.App, c, env)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(e.App),
			},
		})

		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		scheduler := cron.New()

		// every 20 hours
		scheduler.MustAdd("refresh-token", "* */20 * * *", func() {
			handleRefreshToken(e.App, env)
		})

		scheduler.Start()

		return nil
	})

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/platforms/tiktok/revoke-token",
			Handler: func(c echo.Context) error {
				return handleRevokeToken(e.App, c, env)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(e.App),
			},
		})

		return nil
	})
}
