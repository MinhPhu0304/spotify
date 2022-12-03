package service

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

func (s *Service) RecentTracks(ctx context.Context, spotifyToken string) ([]spotify.RecentlyPlayedItem, error) {
	if spotifyToken == "" {
		return []spotify.RecentlyPlayedItem{}, errors.New("missing spotify token")
	}

	t, err := s.c.RecentTracks(ctx, spotifyToken)
	return t, err
}

func (s *Service) TopTracks(ctx context.Context, spotifyToken string) ([]spotify.FullTrack, error) {
	if spotifyToken == "" {
		return []spotify.FullTrack{}, errors.New("missing spotify token")
	}

	t, err := s.c.TopTracks(ctx, spotifyToken, spotify.Limit(50))
	return t, err
}
