package platforms

import (
	"basedpocket/extension"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/core"
)

type TiktokOAuth2 struct {
}

func (oath2 *TiktokOAuth2) RedirectToOAuth(app core.App, ctx echo.Context, env *extension.Env) error {

	return ctx.String(http.StatusOK, "Hello, World!")
}

func (oath2 *TiktokOAuth2) FetchAccessToken(app core.App, ctx echo.Context, env *extension.Env) error {

	return ctx.String(http.StatusOK, "Hello, World!")
}

func (oath2 *TiktokOAuth2) RefreshAccessToken(app core.App, ctx echo.Context, env *extension.Env) error {

	return ctx.String(http.StatusOK, "Hello, World!")
}

func (oath2 *TiktokOAuth2) RevokeAccessToken(app core.App, ctx echo.Context, env *extension.Env) error {

	return ctx.String(http.StatusOK, "Hello, World!")
}
