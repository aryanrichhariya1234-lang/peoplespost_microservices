package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	AUTH_SERVICE_URL string
	POST_SERVICE_URL string
	AI_SERVICE_URL   string
	PORT             string
)

func LoadEnv() {
	_ = godotenv.Load()

	// ✅ Correct fallbacks (match your architecture)
	AUTH_SERVICE_URL = getEnv("AUTH_SERVICE_URL", "http://auth64000")
	POST_SERVICE_URL = getEnv("POST_SERVICE_URL", "http://posts:5000")
	AI_SERVICE_URL   = getEnv("AI_SERVICE_URL", "http://ai:4001")

	PORT = getEnv("PORT", "4000")

	log.Println("Config Loaded:")
	log.Println("AUTH_SERVICE_URL:", AUTH_SERVICE_URL)
	log.Println("POST_SERVICE_URL:", POST_SERVICE_URL)
	log.Println("AI_SERVICE_URL:", AI_SERVICE_URL)
	log.Println("PORT:", PORT)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}