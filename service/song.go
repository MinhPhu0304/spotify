package service

import (
	"context"
	"sync"
	"time"

	"github.com/MinhPhu0304/spotify/types"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	spotifyClient "github.com/zmb3/spotify/v2"
)

func (s *Service) SongDetails(ctx context.Context, spotifyToken string, trackID string) (types.Song, error) {
	if spotifyToken == "" {
		return types.Song{}, errors.New("missing spotify token")
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	rec := make([]spotifyClient.SimpleTrack, 0)
	wg.Add(1)
	recCtx, recCancel := context.WithTimeout(ctx, 2*time.Second)
	defer recCancel()
	go func(ctx context.Context) {
		defer wg.Done()
		defer sentry.RecoverWithContext(ctx)
		r, err := s.c.Recommendation(ctx, spotifyToken, trackID)
		if err != nil {
			sentry.CaptureException(err)
		}
		defer mu.Unlock()
		mu.Lock()
		rec = r
	}(recCtx)

	feats := spotifyClient.AudioFeatures{}
	wg.Add(1)
	featurectx, featureCancel := context.WithTimeout(ctx, 2*time.Second)
	defer featureCancel()
	go func(ctx context.Context) {
		defer wg.Done()
		defer sentry.RecoverWithContext(ctx)
		f, err := s.c.TrackFeatures(ctx, spotifyToken, []spotifyClient.ID{spotifyClient.ID(trackID)})
		if err != nil {
			sentry.CaptureException(err)
			return
		}
		defer mu.Unlock()
		mu.Lock()
		feats = f[trackID]
	}(featurectx)

	track := spotifyClient.FullTrack{}
	wg.Add(1)
	trackCtx, trackCancel := context.WithTimeout(ctx, 2*time.Second)
	defer trackCancel()
	go func(ctx context.Context) {
		defer wg.Done()
		defer sentry.RecoverWithContext(ctx)
		t, err := s.c.Track(ctx, spotifyToken, trackID)
		if err != nil {
			sentry.CaptureException(err)
			return
		}
		defer mu.Unlock()
		mu.Lock()
		track = t
	}(trackCtx)

	wg.Wait()
	return types.Song{
		Detail:          track,
		Features:        feats,
		Recommendations: rec,
	}, nil
}
