package models

import (
	"basedpocket/extension"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func LoadModels(app *pocketbase.PocketBase, env *extension.Env) {

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// ===================
		// collections
		createCustomersCollection(e.App)
		createChannelCollection(e.App)
		createOAuth2Collection(e.App)
		createPlatformActivityCollection(e.App)

		return nil
	})
}
