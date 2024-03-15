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

const channels string = "channels"

var _ models.Model = (*Channel)(nil)

type Channel struct {
	models.BaseModel
	User            string          `db:"user" json:"user"`
	PlatformName    string          `db:"platform_name" json:"platform_name"`
	ExternalID      string          `db:"external_id" json:"external_id"`
	AccessExpiresIn *types.DateTime `db:"access_expires_in" json:"access_expires_in"`
}
type FindChannelParams struct {
	Id         string `db:"id"`
	User       string `db:"user"`
	ExternalID string `db:"external_id"`
}

func (m *Channel) TableName() string {
	return channels // the name of your collection
}

// ===================================

func (channel *Channel) FindChannel(app core.App, params *FindChannelParams) *utils.CError {

	query := dbx.HashExp{}
	if params.Id != "" {
		query["id"] = params.Id
	}
	if params.User != "" {
		query["user"] = params.User
	}
	if params.ExternalID != "" {
		query["external_id"] = params.ExternalID
	}

	err := app.Dao().ModelQuery(&Channel{}).
		AndWhere(query).
		Limit(1).
		One(channel)

	if err != nil {
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	return nil
}
func (channel *Channel) SaveChannel(app core.App) *utils.CError {
	if err := app.Dao().Save(channel); err != nil {
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	return nil
}

// ===================================

func createChannelCollection(app core.App) {

	collectionName := channels

	existingCollection, _ := app.Dao().FindCollectionByNameOrId(collectionName)
	if existingCollection != nil {
		return
	}

	users, err := app.Dao().FindCollectionByNameOrId(users)
	if err != nil {
		log.Fatalf("users table not found: %+v", err)
	}

	collection := &models.Collection{
		Name:       collectionName,
		Type:       models.CollectionTypeBase,
		ListRule:   types.Pointer("user.id = @request.auth.id"),
		ViewRule:   types.Pointer("user.id = @request.auth.id"),
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
				Name:     "platform_name",
				Type:     schema.FieldTypeText,
				Required: true,
				Options:  &schema.TextOptions{},
			},
			&schema.SchemaField{
				Name:     "external_id",
				Type:     schema.FieldTypeText,
				Required: true,
				Options:  &schema.TextOptions{},
			},
			&schema.SchemaField{
				Name:     "access_expires_in",
				Type:     schema.FieldTypeDate,
				Required: false,
				Options:  &schema.TextOptions{},
			},
		),
		Indexes: types.JsonArray[string]{
			fmt.Sprintf("CREATE UNIQUE INDEX idx_user ON %s (user)", collectionName),
			fmt.Sprintf("CREATE UNIQUE INDEX idx_external_id ON %s (external_id)", collectionName),
			"UNIQUE(external_id)",
		},
	}

	if err := app.Dao().SaveCollection(collection); err != nil {
		log.Fatalf("%s collection failed: %+v", collectionName, err)
	}
}
