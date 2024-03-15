package main

import (
	"basedpocket/base"
	"basedpocket/services/payment"
	"basedpocket/services/tiktok"
	"log"

	"github.com/pocketbase/pocketbase"
)

// go run main.go serve
//

func main() {
	env := base.LoadEnv()
	base.LoadLogging(env)
	app := pocketbase.New()

	payment.LoadPayment(app, env)
	tiktok.LoadTiktok(app, env)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
