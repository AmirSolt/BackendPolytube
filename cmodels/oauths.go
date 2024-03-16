package cmodels

import (
	"fmt"
	"log"

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

func (m *OAuth) TableName() string {
	return oauths // the name of your collection
}

// ============================================

func createOAuthCollection(app core.App) {

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
