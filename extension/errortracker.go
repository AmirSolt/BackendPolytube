package extension

import (
	"log"

	"github.com/getsentry/sentry-go"
)

type AppError struct {
	Message string `json:"message"`
	EventID string `json:"eventID"`
}

func LoadLogging(env *Env) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              env.GLITCHTIP_DSN,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		Debug:            !env.IS_PROD,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	// base.Engine.Use(SentryGinNew(SentryGinOptions{}))

}
