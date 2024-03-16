package cmodels

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

type EventParams struct {
	PlatformAccountID string
	ExternalAccountID string
	Title             string
	Message           string
	Status            EventStatus
}

type EventStatus string

const PrimaryStatus EventStatus = "primary"
const SecondaryStatus EventStatus = "secondary"
const TertiaryStatus EventStatus = "tertiary"
const SuccessStatus EventStatus = "success"
const WarningStatus EventStatus = "warning"
const ErrorStatus EventStatus = "error"
const SurfaceStatus EventStatus = "surface"

// ===================================
// ===================================
// ===================================

const events string = "events"

var _ models.Model = (*Event)(nil)

type Event struct {
	models.BaseModel
	User    string `db:"user" json:"user"`
	Channel string `db:"channel" json:"channel"`
	Message string `db:"message" json:"message"`
	Status  string `db:"status" json:"status"`
}
type FindEventParams struct {
	Id      string `db:"id"`
	User    string `db:"user"`
	Channel string `db:"channel"`
}

func (m *Event) TableName() string {
	return events
}

// ============================================

func createEventCollection(app core.App) {

	collectionName := events

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
