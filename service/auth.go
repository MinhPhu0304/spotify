package service

import (
	"context"
	"net/http"
)

func (s *Service) CompleteAuth(ctx context.Context, r *http.Request) (redirectURI string, err error) {
	return s.c.CompleteAuth(ctx, r)
}

func (s *Service) AuthURL() string {
	return s.c.GetAuthURL()
}
