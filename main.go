package main

import (
	"basedpocket/extension"
	"basedpocket/extension/payment"
	"basedpocket/extension/platforms"
	"log"

	"github.com/pocketbase/pocketbase"
)

// go run main.go serve
//

func main() {

	env := extension.LoadEnv()
	extension.LoadLogging(env)

	app := pocketbase.New()

	payment.LoadPayment(app, env)
	platforms.LoadPlatforms(app, env)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
