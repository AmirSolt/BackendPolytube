package extension

import (
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase/core"
)

type OAuth2 interface {
	RedirectToOAuth(app core.App, ctx echo.Context, env *Env) error
	FetchAccessToken(app core.App, ctx echo.Context, env *Env) error
	RefreshAccessToken(app core.App, ctx echo.Context, env *Env) error
	RevokeAccessToken(app core.App, ctx echo.Context, env *Env) error
}
