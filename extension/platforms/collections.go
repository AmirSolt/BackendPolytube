package platforms

import (
	"fmt"
	"log"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
)

func createPlatformCollection(app core.App) {

	collectionName := "platform_accounts"

	existingCollection, _ := app.Dao().FindCollectionByNameOrId(collectionName)
	if existingCollection != nil {
		return
	}

	users, err := app.Dao().FindCollectionByNameOrId("users")
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
				Name:     "external_account_id",
				Type:     schema.FieldTypeText,
				Required: true,
				Options:  &schema.TextOptions{},
			},
			&schema.SchemaField{
				Name:     "access_can_expire",
				Type:     schema.FieldTypeBool,
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
			fmt.Sprintf("CREATE UNIQUE INDEX idx_account_id ON %s (account_id)", collectionName),
			fmt.Sprintf("UNIQUE(external_account_id)"),
		},
	}

	if err := app.Dao().SaveCollection(collection); err != nil {
		log.Fatalf("%s collection failed: %+v", collectionName, err)
	}
}

func createPlatformActivityCollection(app core.App) {

	collectionName := "platform_activities"

	existingCollection, _ := app.Dao().FindCollectionByNameOrId(collectionName)
	if existingCollection != nil {
		return
	}

	users, err := app.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		log.Fatalf("users table not found: %+v", err)
	}

	platformAccounts, err := app.Dao().FindCollectionByNameOrId("platform_accounts")
	if err != nil {
		log.Fatalf("platform_accounts table not found: %+v", err)
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
				Name:     "platform_account",
				Type:     schema.FieldTypeRelation,
				Required: true,
				Options: &schema.RelationOptions{
					MaxSelect:     types.Pointer(1),
					CollectionId:  platformAccounts.Id,
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

func createOAuth2Collection(app core.App) {

	// OAuth 2.0
	collectionName := "oauths"

	existingCollection, _ := app.Dao().FindCollectionByNameOrId(collectionName)
	if existingCollection != nil {
		return
	}

	users, err := app.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		log.Fatalf("users table not found: %+v", err)
	}

	platformAccounts, err := app.Dao().FindCollectionByNameOrId("platform_accounts")
	if err != nil {
		log.Fatalf("platform_accounts table not found: %+v", err)
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
				Name:     "platform_account",
				Type:     schema.FieldTypeRelation,
				Required: true,
				Options: &schema.RelationOptions{
					MaxSelect:     types.Pointer(1),
					CollectionId:  platformAccounts.Id,
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
