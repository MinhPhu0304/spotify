package spotify

import (
	"context"

	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

func (c *Spotify) Recommendation(ctx context.Context, token string, trackID string) ([]spotify.SimpleTrack, error) {
	client := c.clientWithTrace(ctx, token)
	result, err := client.GetRecommendations(ctx,
		spotify.Seeds{Tracks: []spotify.ID{spotify.ID(trackID)}},
		nil,
		spotify.Limit(100))
	return result.Tracks, errors.Wrap(err, "fail to get spotify recommendation")
}
