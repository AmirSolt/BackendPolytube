package platforms

import (
	"basedpocket/extension"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

func LoadPlatforms(app *pocketbase.PocketBase, env *extension.Env) {
	loadTiktok(app, env)
}

func loadTiktok(app *pocketbase.PocketBase, env *extension.Env) {

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// ===================
		// collections
		createPlatformCollection(e.App)
		createOAuth2Collection(e.App)
		createPlatformActivityCollection(e.App)

		// ===================
		// routes
		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/platforms/tiktok/oauth-request",
			Handler: func(c echo.Context) error {
				return handleOAuthRequest(e.App, c, env)
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
				return handleOAuthSuccess(e.App, c, env)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(e.App),
				apis.RequireRecordAuth("users"),
			},
		})

		e.Router.AddRoute(echo.Route{
			Method: http.MethodPost,
			Path:   "/platforms/tiktok/revoke-token",
			Handler: func(c echo.Context) error {
				return handleRevokeToken(e.App, c, env)
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(e.App),
				apis.RequireRecordAuth("users"),
			},
		})

		return nil
	})
}
