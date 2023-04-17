package service

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"

	"github.com/MinhPhu0304/spotify/repository"
	"github.com/MinhPhu0304/spotify/types"
)

func (s *Service) TopArtists(ctx context.Context, spotifyToken string) ([]spotify.FullArtist, error) {
	if t, err := s.repo.GetTopArtists(spotifyToken); errors.Is(repository.ErrInvalidType, err) || errors.Is(repository.ErrNotFound, err) {
		a, err := s.spotifyClient.TopArtists(ctx, spotifyToken, 50)
		go s.repo.InsertTopArtist(spotifyToken, a)
		return a, err
	} else {
		return t, nil
	}
}

func (s *Service) Artist(ctx context.Context, token string, artistID string) (types.ArtistInfo, error) {
	if a, err := s.repo.GetArtistInfo(artistID); errors.Is(repository.ErrInvalidType, err) || errors.Is(repository.ErrNotFound, err) {
		info, err := s.spotifyClient.Artist(ctx, token, artistID, s.lastFMClient)
		go s.repo.InsertArtistInfo(&info)
		return info, err
	} else {
		return *a, nil
	}
}

func (s *Service) RelatedArtist(ctx context.Context, token string, artistID string) ([]spotify.FullArtist, error) {
	a, err := s.spotifyClient.RelatedArtist(ctx, token, artistID)
	return a, err
}
