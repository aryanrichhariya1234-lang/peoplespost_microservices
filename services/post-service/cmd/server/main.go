package main

import (
	"log"
	"net/http"
	"os"

	"post-service/internal/config"
	"post-service/internal/routes"
)

func main() {

	config.LoadEnv()
	config.ConnectMongo()
	config.InitCloudinary()
	router := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Println("Post service running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}