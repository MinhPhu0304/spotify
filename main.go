package main

import (
	"net/http"
	"os"
	"time"

	"github.com/MinhPhu0304/spotify/routes"
	"github.com/akrylysov/algnhsa"
	"github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"

	log "github.com/sirupsen/logrus"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Warn(err)
	}

	sentrySyncTransport := sentry.NewHTTPSyncTransport()
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		AttachStacktrace: true,
		Environment:      "production",
		Transport:        sentrySyncTransport,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	}); err != nil {
		log.Errorf("sentry.Init: %s", err)
		os.Exit(1)
	}

	isAWS := os.Getenv("AWS_REGION")
	srvCfg := routes.Config{
		SpotifyCallBackURI:  os.Getenv("CALLBACK_URI"),
		SpotifyDashboardURI: os.Getenv("DASHBOARD_URI"),
		LastFMToken:         os.Getenv("LASTFM_API_KEY"),
	}
	server := routes.CreateServer(srvCfg)

	if isAWS != "" {
		//only if production:
		log.Println("Started Lambda function")
		algnhsa.ListenAndServe(server.Handler, nil)
	} else {
		//only if dev
		log.Println("Serving on localhost:8080")
		http.ListenAndServe(":8080", server.Handler)
	}

	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)
}
