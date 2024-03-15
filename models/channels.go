package models

import (
	"basedpocket/utils"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/getsentry/sentry-go"
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
type InsertChannelParams struct {
	User            string          `db:"user"`
	PlatformName    string          `db:"platform_name"`
	ExternalID      string          `db:"external_id"`
	AccessExpiresIn *types.DateTime `db:"access_expires_in"`
}
type GetChannelParams struct {
	Id         string `db:"id"`
	User       string `db:"user"`
	ExternalID string `db:"external_id"`
}

func (m *Channel) TableName() string {
	return channels // the name of your collection
}

// ===================================

func InsertChannel(app core.App, params *InsertChannelParams) (*Channel, *utils.CError) {
	record, err := insertRecord(app, params, "channels")
	if err != nil {
		return nil, err
	}
	return record.(*Channel), nil
}

// ===================================

// Identity params include all parameters that you could/should use to fetch the record

func GetChannelQueryStr(params *GetChannelParams) (string, *utils.CError) {
	var queries []string

	if params.User == "" {
		err := fmt.Errorf("params.User cannot be empty")
		eventID := sentry.CaptureException(err)
		return "", &utils.CError{Message: "Internal Server Error", EventID: fmt.Sprintf("%v", &eventID)}
	}
	tag, errTag := utils.GetFieldTag(params, "User", "db")
	if errTag != nil {
		return "", errTag
	}
	queries = append(queries, fmt.Sprintf("%s=%s", tag, params.User))

	// =======================

	v := reflect.ValueOf(*params)
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Check for non-empty value (excluding zero values for numeric types)

		if !utils.isEmptyValue(value) {
			// Get the db tag (default to field name)
			dbName := field.Tag.Get("db")
			if dbName == "" {
				dbName = field.Name
			}

			// Handle string values for proper quoting
			if value.Kind() == reflect.String {
				quotedValue := fmt.Sprintf("'%s'", value.String())
				queries = append(queries, fmt.Sprintf("%s=%s", dbName, quotedValue))
			} else {
				queries = append(queries, fmt.Sprintf("%s=%v", dbName, value))
			}
		}
	}

	return strings.Join(queries, " AND ")
}

// ===================================

func createChannelCollection(app core.App) {

	collectionName := channels

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
