package service

import (
	"context"

	"github.com/MinhPhu0304/spotify/types"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

func (s *Service) TopArtists(ctx context.Context, spotifyToken string) ([]spotify.FullArtist, error) {
	if spotifyToken == "" {
		return []spotify.FullArtist{}, errors.New("missing spotify token")
	}

	a, err := s.c.TopArtists(ctx, spotifyToken, 50)
	return a, err
}

func (s *Service) Artist(ctx context.Context, token string, artistID string) (types.ArtistInfo, error) {
	if token == "" {
		return types.ArtistInfo{}, errors.New("missing spotify token")
	}

	info, err := s.c.Artist(ctx, token, artistID)
	return info, err
}

func (s *Service) RelatedArtist(ctx context.Context, token string, artistID string) ([]spotify.FullArtist, error) {
	if token == "" {
		return []spotify.FullArtist{}, errors.New("missing spotify token")
	}
	a, err := s.c.RelatedArtist(ctx, token, artistID)
	return a, err
}
