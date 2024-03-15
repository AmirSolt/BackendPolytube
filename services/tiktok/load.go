package tiktok

import (
	"basedpocket/extension"
	"basedpocket/extension/platforms/tiktok"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func LoadTiktok(app *pocketbase.PocketBase, env *extension.Env) {

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// ===================
		// routes
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/platforms/tiktok/oauth-request",
			Handler: func(c echo.Context) error {
				return tiktok.HandleOAuthRequest(e.App, c, env)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(e.App),
				apis.RequireRecordAuth("users"),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/platforms/tiktok/oauth-success",
			Handler: func(c echo.Context) error {
				return tiktok.HandleOAuthSuccess(e.App, c, env)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(e.App),
				apis.RequireRecordAuth("users"),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/platforms/tiktok/:platform_account_id/revoke-token",
			Handler: func(c echo.Context) error {
				return tiktok.HandleRevokeToken(e.App, c, env)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(e.App),
				apis.RequireRecordAuth("users"),
			},
		})

		return nil
	})
}
