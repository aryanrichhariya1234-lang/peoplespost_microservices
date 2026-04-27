package config

import (
	"log"
	"os"
	"github.com/joho/godotenv"
)
var POST_SERVICE_URL string
func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		err = godotenv.Load("../.env")
	}
	if err != nil {
		err = godotenv.Load("../../.env")
	}
	if err != nil {
		log.Println("No .env file found")
	}
	POST_SERVICE_URL = os.Getenv("POST_SERVICE_URL")
	log.Printf(POST_SERVICE_URL)
}