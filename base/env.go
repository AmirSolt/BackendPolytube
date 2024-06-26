package base

import (
	"fmt"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Env struct {
	DOMAIN          string `validate:"url"`
	FRONTEND_DOMAIN string `validate:"url"`
	IS_PROD         bool   `validate:"boolean"`

	STRIPE_PUBLIC_KEY  string `validate:"required"`
	STRIPE_PRIVATE_KEY string `validate:"required"`
	STRIPE_WEBHOOK_KEY string `validate:"required"`

	TIKTOK_CLIENT_KEY    string `validate:"required"`
	TIKTOK_CLIENT_SECRET string `validate:"required"`

	ELEVENLABS_API_KEY string `validate:"required"`

	GLITCHTIP_DSN string `validate:"required"`
}

func LoadEnv() *Env {

	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Warrning .env does not exist:", err)
	}

	env := Env{
		DOMAIN:               os.Getenv("DOMAIN"),
		FRONTEND_DOMAIN:      os.Getenv("FRONTEND_DOMAIN"),
		IS_PROD:              strToBool(os.Getenv("IS_PROD")),
		STRIPE_PUBLIC_KEY:    os.Getenv("STRIPE_PUBLIC_KEY"),
		STRIPE_PRIVATE_KEY:   os.Getenv("STRIPE_PRIVATE_KEY"),
		STRIPE_WEBHOOK_KEY:   os.Getenv("STRIPE_WEBHOOK_KEY"),
		TIKTOK_CLIENT_KEY:    os.Getenv("TIKTOK_CLIENT_KEY"),
		TIKTOK_CLIENT_SECRET: os.Getenv("TIKTOK_CLIENT_SECRET"),
		ELEVENLABS_API_KEY:   os.Getenv("ELEVENLABS_API_KEY"),
		GLITCHTIP_DSN:        os.Getenv("GLITCHTIP_DSN"),
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(env)
	if err != nil {
		log.Fatal("Error .env:", err)
	}

	return &env
}

func strToBool(s string) bool {
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	log.Fatal("Error .env: strToBool failed. string: ", s)
	return false
}
