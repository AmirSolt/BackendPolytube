package cmodels

import (
	"basedpocket/utils"
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

const users string = "users"

var _ models.Model = (*User)(nil)

type User struct {
	models.BaseModel
	Email string `db:"email" json:"email"`
}
type FindUserParams struct {
	Id    string `db:"id"`
	Email string `db:"email"`
}

func (m *User) TableName() string {
	return users // the name of your collection
}

// ===================================

func (user *User) FindUser(app core.App, params *FindUserParams) *utils.CError {

	query := dbx.HashExp{}
	if params.Id != "" {
		query["id"] = params.Id
	}
	if params.Email != "" {
		query["email"] = params.Email
	}
	err := app.Dao().ModelQuery(&User{}).
		AndWhere(query).
		Limit(1).
		One(user)

	if err != nil {
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	return nil
}
func (user *User) SaveUser(app core.App) *utils.CError {
	if err := app.Dao().Save(user); err != nil {
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	return nil
}
func (user *User) GetUserByContext(ctx echo.Context) *utils.CError {
	user = ctx.Get(apis.ContextAuthRecordKey).(*User)
	if user == nil {
		err := fmt.Errorf("user not found")
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	return nil
}
