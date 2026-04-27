package main

import (
	"log"
	"net/http"
	"os"

	"ai-service/internal/config"
	"ai-service/internal/routes"
)

func main() {

	config.LoadEnv()


	router := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "6000"
	}

	log.Println("AI service running on port", port)

	
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}