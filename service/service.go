package service

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/MinhPhu0304/spotify/repository"
	"github.com/MinhPhu0304/spotify/trace"
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
	repo        repository.Repository
}

var authScope = []string{
	spotifyauth.ScopeUserTopRead,
	spotifyauth.ScopeUserFollowRead,
	spotifyauth.ScopeUserReadPrivate,
	spotifyauth.ScopeUserReadRecentlyPlayed,
}

func CreateService(redirectURI string, state string, logger *logrus.Logger, repository repository.Repository) *Service {
	auth := spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(authScope...))

	return &Service{
		spotifyAuth: auth,
		state:       state,
		log:         logger,
		repo:        repository,
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

	client := spotify.New(trace.WrapWithTrace(s.spotifyAuth.Client(ctx, &oauth2.Token{AccessToken: spotifyToken})))
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

	client := spotify.New(trace.WrapWithTrace(s.spotifyAuth.Client(ctx, &oauth2.Token{AccessToken: spotifyToken})))
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

	client := spotify.New(trace.WrapWithTrace(s.spotifyAuth.Client(ctx, &oauth2.Token{AccessToken: spotifyToken})))
	result, err := client.PlayerRecentlyPlayedOpt(ctx, &spotify.RecentlyPlayedOptions{Limit: 50})
	if err != nil {
		return []spotify.RecentlyPlayedItem{}, errors.Wrap(err, "fail to get spotify top tracks")
	}

	return result, nil
}

func (s *Service) RelatedArtist(ctx context.Context, token string, artistID string) ([]spotify.FullArtist, error) {
	if token == "" {
		return []spotify.FullArtist{}, errors.New("missing spotify token")
	}

	client := spotify.New(trace.WrapWithTrace(s.spotifyAuth.Client(ctx, &oauth2.Token{AccessToken: token})))
	artist, err := client.GetRelatedArtists(ctx, spotify.ID(artistID))
	return artist, err
}

func allTrackID(tracks []spotify.FullTrack) []spotify.ID {
	trackIDs := make([]spotify.ID, 0)
	for _, track := range tracks {
		trackIDs = append(trackIDs, track.ID)
	}
	return trackIDs
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
