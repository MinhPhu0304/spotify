package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/zmb3/spotify/v2"

	"github.com/MinhPhu0304/spotify/repository"
	"github.com/MinhPhu0304/spotify/types"
)

func (s *Service) SongDetails(ctx context.Context, spotifyToken string, trackID string) (types.Song, error) {
	if s, _ := s.repo.GetSong(trackID); s != nil {
		return *s, nil
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	rec := make([]spotify.SimpleTrack, 0)
	wg.Add(1)
	recCtx, recCancel := context.WithTimeout(ctx, 2*time.Second)
	defer recCancel()
	go func(ctx context.Context) {
		defer wg.Done()
		defer sentry.RecoverWithContext(ctx)
		r, err := s.spotifyClient.Recommendation(ctx, spotifyToken, trackID)
		if err != nil {
			sentry.CaptureException(err)
		}
		defer mu.Unlock()
		mu.Lock()
		rec = r
	}(recCtx)

	feats := spotify.AudioFeatures{}
	wg.Add(1)
	featurectx, featureCancel := context.WithTimeout(ctx, 2*time.Second)
	defer featureCancel()
	go func(ctx context.Context) {
		defer wg.Done()
		if f := s.getTrackFeatures(ctx, spotifyToken, trackID); f != nil {
			mu.Lock()
			feats = f[trackID]
			mu.Unlock()
		}
	}(featurectx)

	track := spotify.FullTrack{}
	wg.Add(1)
	trackCtx, trackCancel := context.WithTimeout(ctx, 2*time.Second)
	defer trackCancel()
	go func(ctx context.Context) {
		defer wg.Done()
		if t := s.getTrack(ctx, spotifyToken, trackID); t != nil {
			mu.Lock()
			track = *t
			mu.Unlock()
		}
	}(trackCtx)

	wg.Wait()
	song := types.Song{
		Detail:          track,
		Features:        feats,
		Recommendations: rec,
	}
	go s.repo.InsertSong(&song)
	return song, nil
}

func (s *Service) getTrack(ctx context.Context, spotifyToken string, trackID string) *spotify.FullTrack {
	defer sentry.RecoverWithContext(ctx)
	if t, err := s.repo.GetSpotifyFullTrack(trackID); errors.Is(repository.ErrInvalidType, err) || errors.Is(repository.ErrNotFound, err) {
		t, err := s.spotifyClient.Track(ctx, spotifyToken, trackID)
		if err != nil {
			sentry.CaptureException(err)
			return nil
		}
		go s.repo.InsertSpotifyFullTrack(&t)
		return &t
	} else {
		return t
	}
}

func (s *Service) getTrackFeatures(ctx context.Context, spotifyToken string, trackID string) map[string]spotify.AudioFeatures {
	defer sentry.RecoverWithContext(ctx)
	f, err := s.spotifyClient.TrackFeatures(ctx, spotifyToken, []spotify.ID{spotify.ID(trackID)})
	if err != nil {
		sentry.CaptureException(err)
		return f
	}
	return nil
}
