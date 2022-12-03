package spotify

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MinhPhu0304/spotify/repository"
	"github.com/MinhPhu0304/spotify/trace"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type Spotify struct {
	spotifyAuth  *spotifyauth.Authenticator
	state        string
	repo         repository.Repository
	dashboardURI string
}

var authScope = []string{
	spotifyauth.ScopeUserTopRead,
	spotifyauth.ScopeUserFollowRead,
	spotifyauth.ScopeUserReadPrivate,
	spotifyauth.ScopeUserReadRecentlyPlayed,
}

func NewSpotifyClient(redirectURI string, state string, repository repository.Repository, dashboardURI string) *Spotify {
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(authScope...))

	return &Spotify{
		spotifyAuth:  auth,
		state:        state,
		repo:         repository,
		dashboardURI: dashboardURI,
	}
}

func (s *Spotify) CompleteAuth(ctx context.Context, r *http.Request) (redirectURI string, err error) {
	tok, err := s.spotifyAuth.Token(ctx, s.state, r)
	if err != nil {
		return "", errors.Wrap(err, "failed to get token")
	}
	if st := r.FormValue("state"); st != s.state {
		return "", errors.New(fmt.Sprintf("State mismatch: %s != %s", st, s.state))
	}
	dashboardURI := fmt.Sprintf("%s/callback?token=%s&expr=%s", s.dashboardURI, tok.AccessToken, tok.Expiry.UTC().String())
	return dashboardURI, nil
}

func (s *Spotify) GetAuthURL() string {
	return s.spotifyAuth.AuthURL(s.state)
}

func (s *Spotify) TopArtists(ctx context.Context, token string, limit int) ([]spotify.FullArtist, error) {
	client := s.clientWithTrace(ctx, token)
	result, err := client.CurrentUsersTopArtists(ctx, spotify.Limit(limit))
	if err != nil {
		return []spotify.FullArtist{}, errors.Wrap(err, "fail to get spotify top artist")
	}
	return result.Artists, nil
}

func (s *Spotify) TopTracks(ctx context.Context, token string, opts ...spotify.RequestOption) ([]spotify.FullTrack, error) {
	if t, err := s.repo.GetUserTopTracks(token); err != nil {
		return t, nil
	}

	client := s.clientWithTrace(ctx, token)
	result, err := client.CurrentUsersTopTracks(ctx, opts...)
	if err != nil {
		return []spotify.FullTrack{}, errors.Wrap(err, "fail to get spotify top tracks")
	}
	defer func() {
		s.repo.InsertUserTopTracks(result.Tracks, token, nil)
	}()
	return result.Tracks, nil
}

func (s *Spotify) RecentTracks(ctx context.Context, token string) ([]spotify.RecentlyPlayedItem, error) {
	client := s.clientWithTrace(ctx, token)
	result, err := client.PlayerRecentlyPlayedOpt(ctx, &spotify.RecentlyPlayedOptions{Limit: 50})
	if err != nil {
		return []spotify.RecentlyPlayedItem{}, errors.Wrap(err, "fail to get spotify top tracks")
	}

	return result, nil
}

func (s *Spotify) RelatedArtist(ctx context.Context, token string, artistID string) ([]spotify.FullArtist, error) {
	client := s.clientWithTrace(ctx, token)
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

func (s *Spotify) clientWithTrace(ctx context.Context, token string) *spotify.Client {
	return spotify.New(trace.WrapWithTrace(s.spotifyAuth.Client(ctx, &oauth2.Token{AccessToken: token})))
}
