package models

import (
	"basedpocket/utils"
	"reflect"

	"github.com/getsentry/sentry-go"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/models"
)

func insertRecord[T any](app core.App, params *T, collectionName string) (*models.Record, *utils.CError) {
	collection, err := app.Dao().FindCollectionByNameOrId(collectionName)
	if err != nil {
		eventID := sentry.CaptureException(err)
		return nil, &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}

	record := models.NewRecord(collection)

	// Leverage reflect for generic field extraction and setting
	v := reflect.ValueOf(params).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		tag := v.Type().Field(i).Tag.Get("db")
		record.Set(tag, f.Interface())
	}

	if err := app.Dao().SaveRecord(record); err != nil {
		eventID := sentry.CaptureException(err)
		return nil, &utils.CError{Message: "Internal Server Error", EventID: *eventID, Error: err}
	}

	return record, nil
}
