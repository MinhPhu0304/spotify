package service

import (
	"github.com/MinhPhu0304/spotify/client/lastfm"
	"github.com/MinhPhu0304/spotify/client/spotify"
	"github.com/MinhPhu0304/spotify/repository"
)

type Service struct {
	spotifyClient *spotify.Spotify
	lastFMClient  lastfm.LastFMClient
	repo          repository.Repository
}

func NewService(spotifyClient *spotify.Spotify, lastFMClient lastfm.LastFMClient, repo repository.Repository) *Service {
	return &Service{
		spotifyClient: spotifyClient,
		lastFMClient:  lastFMClient,
		repo:          repo,
	}
}
