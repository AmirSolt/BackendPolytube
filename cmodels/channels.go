package cmodels

import (
	"fmt"
	"log"

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
	Platform        string          `db:"platform" json:"platform"`
	Language        string          `db:"language" json:"language"`
	AccessExpiresIn *types.DateTime `db:"access_expires_in" json:"access_expires_in"`
}
type FindChannelParams struct {
	Id       string `db:"id"`
	User     string `db:"user"`
	Platform string `db:"platform"`
}

func (m *Channel) TableName() string {
	return channels // the name of your collection
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

	platforms, err := app.Dao().FindCollectionByNameOrId(platforms)
	if err != nil {
		log.Fatalf("users table not found: %+v", err)
	}

	collection := &models.Collection{
		Name:       collectionName,
		Type:       models.CollectionTypeBase,
		ListRule:   types.Pointer("user.id = @request.auth.id"),
		ViewRule:   types.Pointer("user.id = @request.auth.id"),
		CreateRule: types.Pointer("user.id = @request.auth.id"),
		UpdateRule: types.Pointer("user.id = @request.auth.id"),
		DeleteRule: types.Pointer("user.id = @request.auth.id"),
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
				Name:     "platform",
				Type:     schema.FieldTypeRelation,
				Required: true,
				Options: &schema.RelationOptions{
					MaxSelect:     types.Pointer(1),
					CollectionId:  platforms.Id,
					CascadeDelete: true,
				},
			},
			&schema.SchemaField{
				Name:     "language",
				Type:     schema.FieldTypeText,
				Required: true,
				Options:  &schema.TextOptions{},
			},
			&schema.SchemaField{
				Name:     "access_expires_in",
				Type:     schema.FieldTypeDate,
				Required: false,
				Options:  &schema.DateOptions{},
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
