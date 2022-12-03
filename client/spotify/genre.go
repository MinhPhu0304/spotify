package spotify

import (
	"context"

	"github.com/pkg/errors"
)

func (s *Spotify) Genres(ctx context.Context, token string) ([]string, error) {
	if g, err := s.repo.GetGenres(); err != nil && len(g) != 0 {
		return g, nil
	}

	client := s.clientWithTrace(ctx, token)
	result, err := client.GetAvailableGenreSeeds(ctx)
	if err != nil {
		return []string{}, errors.Wrap(err, "fail to get spotify genre seeds")
	}

	defer func() {
		s.repo.InsertGenres(result, nil)
	}()
	return result, nil
}
