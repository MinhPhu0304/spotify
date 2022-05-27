package service

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type Service struct {
	spotifyAuth *spotifyauth.Authenticator
	state       string
	log         *logrus.Logger
}

var authScope = []string{
	spotifyauth.ScopeUserTopRead,
	spotifyauth.ScopeUserFollowRead,
	spotifyauth.ScopeUserReadRecentlyPlayed,
}

func CreateService(redirectURI string, state string, logger *logrus.Logger) *Service {
	auth := spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(authScope...))

	return &Service{
		spotifyAuth: auth,
		state:       state,
		log:         logger,
	}
}

func (s *Service) CompleteAuth(ctx context.Context, r *http.Request) (redirectURI string, err error) {
	tok, err := s.spotifyAuth.Token(ctx, s.state, r)
	if err != nil {
		return "", errors.Wrap(err, "failed to get token")
	}
	if st := r.FormValue("state"); st != s.state {
		return "", errors.New(fmt.Sprintf("State mismatch: %s != %s", st, s.state))
	}

	dashboardURI := os.Getenv("DASHBOARD_URI") + "/callback?token=" + tok.AccessToken + "&expr=" + tok.Expiry.UTC().String()
	return dashboardURI, nil
}

func (s *Service) GetAuthURL() string {
	return s.spotifyAuth.AuthURL(s.state)
}

func (s *Service) TopArtists(ctx context.Context, spotifyToken string) (topArtists []spotify.FullArtist, err error) {
	if spotifyToken == "" {
		return []spotify.FullArtist{}, errors.New("missing spotify token")
	}

	client := spotify.New(s.spotifyAuth.Client(ctx, &oauth2.Token{AccessToken: spotifyToken}))
	result, err := client.CurrentUsersTopArtists(ctx, spotify.Limit(50))
	if err != nil {
		return []spotify.FullArtist{}, errors.Wrap(err, "fail to get spotify top artist")
	}
	if err != nil {
		return []spotify.FullArtist{}, errors.Wrap(err, "fail to get spotify top artist")
	}
	return result.Artists, nil
}

func (s *Service) TopTracks(ctx context.Context, spotifyToken string) (topArtists []spotify.FullTrack, err error) {
	if spotifyToken == "" {
		return []spotify.FullTrack{}, errors.New("missing spotify token")
	}

	client := spotify.New(s.spotifyAuth.Client(ctx, &oauth2.Token{AccessToken: spotifyToken}))
	result, err := client.CurrentUsersTopTracks(ctx, spotify.Limit(50))
	if err != nil {
		return []spotify.FullTrack{}, errors.Wrap(err, "fail to get spotify top tracks")
	}

	return result.Tracks, nil
}

func (s *Service) RecentTracks(ctx context.Context, spotifyToken string) (topArtists []spotify.RecentlyPlayedItem, err error) {
	if spotifyToken == "" {
		return []spotify.RecentlyPlayedItem{}, errors.New("missing spotify token")
	}

	client := spotify.New(s.spotifyAuth.Client(ctx, &oauth2.Token{AccessToken: spotifyToken}))
	result, err := client.PlayerRecentlyPlayedOpt(ctx, &spotify.RecentlyPlayedOptions{Limit: 50})
	if err != nil {
		return []spotify.RecentlyPlayedItem{}, errors.Wrap(err, "fail to get spotify top tracks")
	}

	return result, nil
}

type ArtistInfo struct {
	Artirst   spotify.FullArtist  `json:"artist"`
	TopTracks []spotify.FullTrack `json:"topTracks"`
}

func (s *Service) Artist(ctx context.Context, token string, artistID string) (ArtistInfo, error) {
	if token == "" {
		return ArtistInfo{}, errors.New("missing spotify token")
	}

	client := spotify.New(s.spotifyAuth.Client(ctx, &oauth2.Token{AccessToken: token}))
	artist, err := client.GetArtist(ctx, spotify.ID(artistID))

	if err != nil {
		return ArtistInfo{}, errors.Wrap(err, "failed to get spotify artist")
	}

	topTracks, err := client.GetArtistsTopTracks(ctx, spotify.ID(artistID), "NZ")
	if err != nil {
		return ArtistInfo{}, errors.Wrap(err, "failed to get spotify artist top tracks")
	}

	return ArtistInfo{
		Artirst:   *artist,
		TopTracks: topTracks,
	}, nil
}
