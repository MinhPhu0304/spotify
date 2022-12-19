package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MinhPhu0304/spotify/client/spotify"
	"github.com/MinhPhu0304/spotify/repository"
	"github.com/MinhPhu0304/spotify/service"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	service service.Service
	Handler http.Handler
}

func CreateServer() Server {
	redirectURI := os.Getenv("CALLBACK_URI")
	repo := repository.CreateInMemoryRepo()
	sc := spotify.NewSpotifyClient(redirectURI, "spotifyOauth", repo, os.Getenv("DASHBOARD_URI"))
	srvc := service.NewService(sc)

	// Create an instance of sentryhttp
	sentryHandler := sentryhttp.New(sentryhttp.Options{Repanic: true})
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Second * 60))
	r.Use(cors.AllowAll().Handler)
	s := Server{
		service: *srvc,
		Handler: r,
	}

	// Public route - only needs standard middleware
	r.Group(func(r chi.Router) {
		r.HandleFunc("/callback", sentryHandler.HandleFunc(s.HandleCallback))
		r.HandleFunc("/oauth/spotify", sentryHandler.HandleFunc(s.HandleLoginSpotify))
		r.HandleFunc("/ping", s.HandlePing)
	})

	// Private route must have spotify token
	r.Group(func(r chi.Router) {
		r.Use(MustHaveSpotifyToken())
		r.Get("/personal/top_artists", sentryHandler.HandleFunc(s.HandleTopArtists))
		r.Get("/personal/top_tracks", sentryHandler.HandleFunc(s.HandleTopTracks))
		r.Get("/genres", sentryHandler.HandleFunc(s.HandleGetGenres))
		r.Get("/personal/recently_played", sentryHandler.HandleFunc(s.HandleRecentlyPlayed))
		r.Get("/artist/{id}", sentryHandler.HandleFunc(s.HandleGetArtist))
		r.Get("/song/{id}", sentryHandler.HandleFunc(s.HandleGetSong))
		r.Get("/artist/{id}/related-artists", sentryHandler.HandleFunc(s.HandleGetRelatedArtist))
	})

	return s
}

func (s *Server) HandleCallback(w http.ResponseWriter, r *http.Request) {
	redirectURI, err := s.service.CompleteAuth(r.Context(), r)
	if err != nil {
		http.Error(w, "Failed to login to Spotify", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, redirectURI, http.StatusFound)
}

func (s *Server) HandlePing(w http.ResponseWriter, r *http.Request) {
	http.StatusText(http.StatusOK)
}

func (s *Server) HandleTopArtists(w http.ResponseWriter, r *http.Request) {
	span := sentry.TransactionFromContext(r.Context())
	if span == nil {
		span = sentry.StartSpan(r.Context(), r.Method+" "+r.URL.Path)
	}
	defer span.Finish()
	spotifyToken := r.Header.Get("spotify-token")
	topArtists, err := s.service.TopArtists(r.Context(), spotifyToken)

	if err != nil && (strings.Contains(err.Error(), "The access token expired") || strings.Contains(err.Error(), "Invalid access token")) {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if err != nil {
		HandleError(w, "failed to get top artists", err, http.StatusInternalServerError)
		return
	}

	resBody, err := json.Marshal(topArtists)
	if err != nil {
		HandleError(w, "failed to request to spotify artist", err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resBody)
}

func (s *Server) HandleTopTracks(w http.ResponseWriter, r *http.Request) {
	span := sentry.TransactionFromContext(r.Context())
	if span == nil {
		span = sentry.StartSpan(r.Context(), r.Method+" "+r.URL.Path)
	}
	defer span.Finish()
	spotifyToken := r.Header.Get("spotify-token")
	topTracks, err := s.service.TopTracks(r.Context(), spotifyToken)

	if err != nil && (strings.Contains(err.Error(), "The access token expired") || strings.Contains(err.Error(), "Invalid access token")) {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if err != nil {
		HandleError(w, "failed to get top tracks", err, http.StatusInternalServerError)
		return
	}

	resBody, err := json.Marshal(topTracks)
	if err != nil {
		sentry.CaptureException(err)
		http.Error(w, "failed to request to spotify tracks", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resBody)
}

func (s *Server) HandleRecentlyPlayed(w http.ResponseWriter, r *http.Request) {
	span := sentry.TransactionFromContext(r.Context())
	if span == nil {
		span = sentry.StartSpan(r.Context(), r.Method+" "+r.URL.Path)
	}
	defer span.Finish()
	spotifyToken := r.Header.Get("spotify-token")
	recentlyPlayed, err := s.service.RecentTracks(r.Context(), spotifyToken)

	if err != nil && (strings.Contains(err.Error(), "The access token expired") || strings.Contains(err.Error(), "Invalid access token")) {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if err != nil {
		HandleError(w, "failed to get recently played tracks", err, http.StatusInternalServerError)
		return
	}

	resBody, err := json.Marshal(recentlyPlayed)
	if err != nil {
		sentry.CaptureException(err)
		http.Error(w, "failed to request to recently played tracks", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resBody)
}

func (s *Server) HandleGetArtist(w http.ResponseWriter, r *http.Request) {
	span := sentry.TransactionFromContext(r.Context())
	if span == nil {
		span = sentry.StartSpan(r.Context(), r.Method+" "+"/artist/{id}")
	}
	defer span.Finish()
	spotifyToken := r.Header.Get("spotify-token")
	artistID := chi.URLParam(r, "id")
	artistInfo, err := s.service.Artist(r.Context(), spotifyToken, artistID)

	if err != nil && (strings.Contains(err.Error(), "The access token expired") || strings.Contains(err.Error(), "Invalid access token")) {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if err != nil {
		HandleError(w, "failed to get artist info", err, http.StatusInternalServerError)
		return
	}

	resBody, err := json.Marshal(artistInfo)
	if err != nil {
		HandleError(w, "", err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resBody)
}

func (s *Server) HandleGetRelatedArtist(w http.ResponseWriter, r *http.Request) {
	span := sentry.TransactionFromContext(r.Context())
	if span == nil {
		span = sentry.StartSpan(r.Context(), r.Method+" "+"/artist/{id}/related-artists")
	}
	defer span.Finish()
	spotifyToken := r.Header.Get("spotify-token")
	artistID := chi.URLParam(r, "id")
	artistInfo, err := s.service.RelatedArtist(r.Context(), spotifyToken, artistID)

	if err != nil && (strings.Contains(err.Error(), "The access token expired") || strings.Contains(err.Error(), "Invalid access token")) {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if err != nil {
		HandleError(w, "failed to get related artists", err, http.StatusInternalServerError)
		return
	}

	resBody, err := json.Marshal(artistInfo)
	if err != nil {
		HandleError(w, "", err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resBody)
}

func (s *Server) HandleLoginSpotify(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.service.AuthURL(), http.StatusFound)
}

func (s *Server) HandleGetGenres(w http.ResponseWriter, r *http.Request) {
	span := sentry.TransactionFromContext(r.Context())
	if span == nil {
		span = sentry.StartSpan(r.Context(), r.Method+" "+"/artist/{id}/related-artists")
	}
	defer span.Finish()
	spotifyToken := r.Header.Get("spotify-token")
	g, err := s.service.Genres(r.Context(), spotifyToken)
	if err != nil {
		HandleError(w, "failed to get genres", err, http.StatusInternalServerError)
		return
	}
	resBody, err := json.Marshal(struct {
		Genres []string `json:"genres"`
	}{Genres: g})
	if err != nil {
		HandleError(w, "", err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resBody)
}

func (s *Server) HandleGetSong(w http.ResponseWriter, r *http.Request) {
	span := sentry.TransactionFromContext(r.Context())
	if span == nil {
		span = sentry.StartSpan(r.Context(), r.Method+" "+"/song/{id}")
	}
	defer span.Finish()
	spotifyToken := r.Header.Get("spotify-token")
	artistID := chi.URLParam(r, "id")
	sDetails, err := s.service.SongDetails(r.Context(), spotifyToken, artistID)

	if err != nil && (strings.Contains(err.Error(), "The access token expired") || strings.Contains(err.Error(), "Invalid access token")) {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if err != nil {
		HandleError(w, "failed to get song info", err, http.StatusInternalServerError)
		return
	}

	resBody, err := json.Marshal(sDetails)
	if err != nil {
		HandleError(w, "", err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resBody)
}
