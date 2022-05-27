package main

import (
	"net/http"
	"os"

	"github.com/MinhPhu0304/spotify/routes"
	"github.com/akrylysov/algnhsa"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Warn(err)
	}

	isAWS := os.Getenv("AWS_REGION")
	server := routes.CreateServer()

	if isAWS != "" {
		//only if production:
		log.Println("Started Lambda function")
		algnhsa.ListenAndServe(server.Handler, nil)
	} else {
		//only if dev
		log.Println("Serving on localhost:8080")
		http.ListenAndServe(":8080", server.Handler)
	}
}
