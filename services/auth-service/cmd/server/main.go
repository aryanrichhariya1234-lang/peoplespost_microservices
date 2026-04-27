package main

import (
	"log"
	"net/http"
	"os"

	"auth-service/internal/config"
	"auth-service/internal/routes"
)

func main() {

	config.LoadEnv()
	config.ConnectMongo()

	router := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	log.Println("Auth service running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}