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

	t, err := s.spotifyClient.RecentTracks(ctx, spotifyToken)
	return t, err
}

func (s *Service) TopTracks(ctx context.Context, spotifyToken string) ([]spotify.FullTrack, error) {
	t, err := s.repo.GetUserTopTracks(spotifyToken)

	if err == nil {
		return t, nil
	}

	t, err = s.spotifyClient.TopTracks(ctx, spotifyToken, spotify.Limit(50))
	go s.repo.InsertUserTopTracks(t, spotifyToken)
	return t, err
}
