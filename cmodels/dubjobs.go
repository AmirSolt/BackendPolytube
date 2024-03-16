package cmodels

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

const dubjobs string = "dubjobs"

var _ models.Model = (*Dubjob)(nil)

type Dubjob struct {
	models.BaseModel
	User            string          `db:"user" json:"user"`
	Channel         string          `db:"channel" json:"channel"`
	SourceURL       string          `db:"source_url" json:"source_url"`
	TargetLanguage  string          `db:"target_language" json:"target_language"`
	ExternalID      string          `db:"external_id" json:"external_id"`
	ExpectedReadyIn *types.DateTime `db:"expected_ready_in" json:"expected_ready_in"`
	OutputURL       string          `db:"output_url" json:"output_url"`
	FinishedIn      *types.DateTime `db:"finished_in" json:"finished_in"`
}
type FindDubjobParams struct {
	Id         string `db:"id"`
	User       string `db:"user"`
	ExternalID string `db:"external_id"`
}

func (m *Dubjob) TableName() string {
	return dubjobs // the name of your collection
}

// ============================================

func createDubjobCollection(app core.App) {

	collectionName := dubjobs

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
				Name:     "source_url",
				Type:     schema.FieldTypeUrl,
				Required: true,
				Options:  &schema.TextOptions{},
			},
			&schema.SchemaField{
				Name:     "target_language",
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
				Name:     "expected_ready_in",
				Type:     schema.FieldTypeDate,
				Required: true,
				Options:  &schema.TextOptions{},
			},
			&schema.SchemaField{
				Name:     "output_url",
				Type:     schema.FieldTypeUrl,
				Required: false,
				Options:  &schema.TextOptions{},
			},
			&schema.SchemaField{
				Name:     "finished_in",
				Type:     schema.FieldTypeDate,
				Required: false,
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
