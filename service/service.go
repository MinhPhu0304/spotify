package service

import (
	"github.com/MinhPhu0304/spotify/client/spotify"
)

type Service struct {
	c *spotify.Spotify
}

func NewService(spotifyClient *spotify.Spotify) *Service {
	return &Service{
		c: spotifyClient,
	}
}
