package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MinhPhu0304/spotify/service"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	service service.Service
	Handler http.Handler
}

func CreateServer() Server {
	redirectURI := os.Getenv("CALLBACK_URI")
	logger := log.New()
	service := service.CreateService(redirectURI, "spotifyOauth", logger)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(time.Second * 60))
	r.Use(cors.AllowAll().Handler)
	s := Server{
		service: *service,
		Handler: r,
	}

	// Public route - only needs standard middleware
	r.Group(func(r chi.Router) {
		r.HandleFunc("/callback", s.HandleCallback)
		r.HandleFunc("/oauth/spotify", s.HandleLoginSpotify)
		r.HandleFunc("/ping", s.HandlePing)
	})

	// Private route must have spotify token
	r.Group(func(r chi.Router) {
		r.Use(MustHaveSpotifyToken())
		r.Get("/personal/top_artists", s.HandleTopArtists)
		r.Get("/personal/top_tracks", s.HandleTopTracks)
		r.Get("/personal/recently_played", s.HandleRecentlyPlayed)
		r.Get("/artist/{id}", s.HandleGetArtist)
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
	spotifyToken := r.Header.Get("spotify-token")
	topArtists, err := s.service.TopArtists(r.Context(), spotifyToken)

	if err != nil && strings.Contains(err.Error(), "The access token expired") {
		http.Redirect(w, r, os.Getenv("DASHBOARD_URI"), http.StatusFound)
		return
	}

	if err != nil {
		log.Error(err)
		http.Error(w, "failed to get top artists", http.StatusInternalServerError)
		return
	}

	resBody, err := json.Marshal(topArtists)
	if err != nil {
		log.Error(err)
		http.Error(w, "failed to request to spotify artist", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resBody)
}

func (s *Server) HandleTopTracks(w http.ResponseWriter, r *http.Request) {
	spotifyToken := r.Header.Get("spotify-token")
	topTracks, err := s.service.TopTracks(r.Context(), spotifyToken)
	if err != nil {
		log.WithField("Headers", r.Header).Error(err)
		http.Error(w, "failed to get top tracks", http.StatusInternalServerError)
		return
	}

	resBody, err := json.Marshal(topTracks)
	if err != nil {
		log.Error(err)
		http.Error(w, "failed to request to spotify tracks", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resBody)
}

func (s *Server) HandleRecentlyPlayed(w http.ResponseWriter, r *http.Request) {
	spotifyToken := r.Header.Get("spotify-token")
	recentlyPlayed, err := s.service.RecentTracks(r.Context(), spotifyToken)

	if err != nil && strings.Contains(err.Error(), "The access token expired") {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if err != nil {
		log.WithField("Headers", r.Header).Error(err)
		http.Error(w, "failed to get recently played tracks", http.StatusInternalServerError)
		return
	}

	resBody, err := json.Marshal(recentlyPlayed)
	if err != nil {
		log.Error(err)
		http.Error(w, "failed to request to recently played tracks", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resBody)
}

func (s *Server) HandleGetArtist(w http.ResponseWriter, r *http.Request) {
	spotifyToken := r.Header.Get("spotify-token")
	artistID := chi.URLParam(r, "id")
	log.Info(artistID)
	artistInfo, err := s.service.Artist(r.Context(), spotifyToken, artistID)

	if err != nil && strings.Contains(err.Error(), "The access token expired") {
		http.Redirect(w, r, os.Getenv("DASHBOARD_URI"), http.StatusFound)
		return
	}

	if err != nil {
		log.WithField("Headers", r.Header).Error(err)
		http.Error(w, "failed to get artist info", http.StatusInternalServerError)
		return
	}

	resBody, err := json.Marshal(artistInfo)
	if err != nil {
		log.Error(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resBody)
}

func (s *Server) HandleLoginSpotify(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.service.GetAuthURL(), http.StatusFound)
}
