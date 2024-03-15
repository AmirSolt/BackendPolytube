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

type AcitvityParams struct {
	PlatformAccountID string
	ExternalAccountID string
	Title             string
	Message           string
	Status            ActivityStatus
}

type ActivityStatus string

const PrimaryStatus ActivityStatus = "primary"
const SecondaryStatus ActivityStatus = "secondary"
const TertiaryStatus ActivityStatus = "tertiary"
const SuccessStatus ActivityStatus = "success"
const WarningStatus ActivityStatus = "warning"
const ErrorStatus ActivityStatus = "error"
const SurfaceStatus ActivityStatus = "surface"

// ===================================
// ===================================
// ===================================

const activities string = "activities"

var _ models.Model = (*Activity)(nil)

type Activity struct {
	models.BaseModel
	User    string `db:"user" json:"user"`
	Channel string `db:"channel" json:"channel"`
	Message string `db:"message" json:"message"`
	Status  string `db:"status" json:"status"`
}
type FindActivityParams struct {
	Id      string `db:"id"`
	User    string `db:"user"`
	Channel string `db:"channel"`
}

func (m *Activity) TableName() string {
	return activities // the name of your collection
}

// ===================================

func (activity *Activity) FindActivity(app core.App, params *FindActivityParams) *utils.CError {

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

	err := app.Dao().ModelQuery(&Channel{}).
		AndWhere(query).
		Limit(1).
		One(activity)

	if err != nil {
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	return nil
}
func (activity *Activity) SaveActivity(app core.App) *utils.CError {
	if err := app.Dao().Save(activity); err != nil {
		eventID := sentry.CaptureException(err)
		return &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}
	return nil
}

// ============================================

func createActivityCollection(app core.App) {

	collectionName := activities

	existingCollection, _ := app.Dao().FindCollectionByNameOrId(collectionName)
	if existingCollection != nil {
		return
	}

	users, err := app.Dao().FindCollectionByNameOrId(users)
	if err != nil {
		log.Fatalf("users table not found: %+v", err)
	}

	channels, err := app.Dao().FindCollectionByNameOrId(channels)
	if err != nil {
		log.Fatalf("channels table not found: %+v", err)
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
				Name:     "title",
				Type:     schema.FieldTypeText,
				Required: true,
				Options:  &schema.TextOptions{},
			},
			&schema.SchemaField{
				Name:     "message",
				Type:     schema.FieldTypeText,
				Required: true,
				Options:  &schema.TextOptions{},
			},
			&schema.SchemaField{
				Name:     "status",
				Type:     schema.FieldTypeText,
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
