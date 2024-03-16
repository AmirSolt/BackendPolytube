package cmodels

import (
	"basedpocket/base"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func LoadModels(app *pocketbase.PocketBase, env *base.Env) {

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// ===================
		// collections
		createCustomersCollection(e.App)
		createPlatformCollection(e.App)
		createChannelCollection(e.App)
		createEventCollection(e.App)
		createDubjobCollection(e.App)
		createOAuthCollection(e.App)

		return nil
	})
}
