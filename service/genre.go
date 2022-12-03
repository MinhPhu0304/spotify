package service

import "context"

func (s *Service) Genres(ctx context.Context, token string) ([]string, error) {
	return s.c.Genres(ctx, token)
}
