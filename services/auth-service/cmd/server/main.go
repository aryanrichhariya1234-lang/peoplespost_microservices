package main

import (
	"log"
	"net/http"
	

	"auth-service/internal/config"
	"auth-service/internal/routes"
)

func main() {

	config.LoadEnv()
	config.ConnectMongo()

	router := routes.SetupRouter()

	port :="6000"

	log.Println("Auth service running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}