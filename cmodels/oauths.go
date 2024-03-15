package cmodels

import (
	"basedpocket/utils"
	"fmt"
	"log"

	"github.com/getsentry/sentry-go"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

const oauths string = "oauths"

var _ models.Model = (*OAuth)(nil)

type OAuth struct {
	models.BaseModel
	User                  string          `db:"user" json:"user"`
	Channel               string          `db:"channel" json:"channel"`
	Scope                 string          `db:"scope" json:"scope"`
	AccessToken           string          `db:"access_token" json:"access_token"`
	AccessTokenExpiresIn  *types.DateTime `db:"access_token_expires_in" json:"access_token_expires_in"`
	RefreshToken          string          `db:"refresh_token" json:"refresh_token"`
	RefreshTokenExpiresIn *types.DateTime `db:"refresh_token_expires_in" json:"refresh_token_expires_in"`
}
type FindOAuthParams struct {
	Id           string `db:"id"`
	User         string `db:"user"`
	Channel      string `db:"channel"`
	AccessToken  string `db:"access_token"`
	RefreshToken string `db:"refresh_token"`
}

func (m *OAuth) TableName() string {
	return oauths // the name of your collection
}

// ===================================

func (oauth *OAuth) FindOAuth(app core.App, params *FindOAuthParams) *utils.CError {

	query := dbx.HashExp{}
	if params.Id != "" {
		query["id"] = params.Id
	}
	if params.User != "" {
		query["user"] = params.User
	}
	if params.Channel != "" {
		query["channel"] = params.Channel
	}
	if params.AccessToken != "" {
		query["access_token"] = params.AccessToken
	}
	if params.RefreshToken != "" {
		query["refresh_token"] = params.RefreshToken
	}

	err := app.Dao().ModelQuery(&Channel{}).
		AndWhere(query).
		Limit(1).
		One(oauth)

	if err != nil {
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	return nil
}
func (oauth *OAuth) SaveOAuth(app core.App) *utils.CError {
	if err := app.Dao().Save(oauth); err != nil {
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	return nil
}

// ============================================

func createOAuth2Collection(app core.App) {

	// OAuth 2.0
	collectionName := oauths

	existingCollection, _ := app.Dao().FindCollectionByNameOrId(collectionName)
	if existingCollection != nil {
		return
	}

	users, err := app.Dao().FindCollectionByNameOrId(users)
	if err != nil {
		log.Fatalf("users table not found: %+v", err)
	}

	channels, err := app.Dao().FindCollectionByNameOrId("channels")
	if err != nil {
		log.Fatalf("channels table not found: %+v", err)
	}

	collection := &models.Collection{
		Name:       collectionName,
		Type:       models.CollectionTypeBase,
		ListRule:   nil,
		ViewRule:   nil,
		CreateRule: nil,
		UpdateRule: nil,
		DeleteRule: nil,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:     "user",
				Type:     schema.FieldTypeRelation,
				Required: true,
				Options: &schema.RelationOptions{
					MaxSelect:     types.Pointer(1),
					CollectionId:  users.Id,
					CascadeDelete: true,
				},
			},
			&schema.SchemaField{
				Name:     "channel",
				Type:     schema.FieldTypeRelation,
				Required: true,
				Options: &schema.RelationOptions{
					MaxSelect:     types.Pointer(1),
					CollectionId:  channels.Id,
					CascadeDelete: true,
				},
			},
			&schema.SchemaField{
				Name:     "scope",
				Type:     schema.FieldTypeText,
				Required: true,
				Options:  &schema.TextOptions{},
			},
			&schema.SchemaField{
				Name:     "access_token",
				Type:     schema.FieldTypeText,
				Required: true,
				Options:  &schema.TextOptions{},
			},
			&schema.SchemaField{
				Name:     "access_token_expires_in",
				Type:     schema.FieldTypeDate,
				Required: true,
				Options:  &schema.TextOptions{},
			},
			&schema.SchemaField{
				Name:     "refresh_token",
				Type:     schema.FieldTypeText,
				Required: true,
				Options:  &schema.TextOptions{},
			},
			&schema.SchemaField{
				Name:     "refresh_token_expires_in",
				Type:     schema.FieldTypeDate,
				Required: true,
				Options:  &schema.TextOptions{},
			},
		),
		Indexes: types.JsonArray[string]{
			fmt.Sprintf("CREATE UNIQUE INDEX idx_user ON %s (user)", collectionName),
		},
	}

	if err := app.Dao().SaveCollection(collection); err != nil {
		log.Fatalf("%s collection failed: %+v", collectionName, err)
	}
}
