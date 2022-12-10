package spotify

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

func (c *Spotify) Track(ctx context.Context, token, trackId string) (spotify.FullTrack, error) {
	client := c.clientWithTrace(ctx, token)
	t, err := client.GetTrack(ctx, spotify.ID(trackId))
	return *t, errors.Wrapf(err, "failed to get track detail")
}

func tracksFeatures(ctx context.Context, client *spotify.Client, trackIDs []spotify.ID) (map[string]spotify.AudioFeatures, error) {
	tracksFeatures, err := client.GetAudioFeatures(ctx, trackIDs...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	features := make(map[string]spotify.AudioFeatures)
	for _, track := range tracksFeatures {
		features[string(track.ID)] = *track
	}

	return features, nil
}

func (c *Spotify) TrackFeatures(ctx context.Context, token string, trackIDs []spotify.ID) (map[string]spotify.AudioFeatures, error) {
	sCl := c.clientWithTrace(ctx, token)

	fs, err := tracksFeatures(ctx, sCl, trackIDs)
	return fs, errors.Wrap(err, "failed to get track features")
}
