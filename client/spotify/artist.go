package spotify

import (
	"context"
	"strings"
	"sync"

	"github.com/MinhPhu0304/spotify/client/lastfm"
	"github.com/MinhPhu0304/spotify/repository"
	"github.com/MinhPhu0304/spotify/types"
	lastfmtype "github.com/MinhPhu0304/spotify/types/lastfm"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

func (s *Spotify) artist(ctx context.Context, artistID string, client *spotify.Client) (*spotify.FullArtist, error) {
	if artist, _ := s.repo.GetSpotifyArtist(artistID); artist != nil {
		return artist, nil
	}
	artist, err := client.GetArtist(ctx, spotify.ID(artistID))
	return artist, err
}

func (s *Spotify) Artist(ctx context.Context, token string, artistID string, lastFMClient lastfm.LastFMClient) (types.ArtistInfo, error) {
	spotifyClient := s.clientWithTrace(ctx, token)
	artist, err := s.artist(ctx, artistID, spotifyClient)

	if err != nil {
		return types.ArtistInfo{}, errors.Wrap(err, "failed to get spotify artist")
	}

	bio := lastfmtype.LastFMBio{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer sentry.RecoverWithContext(ctx)
		if abio, err := s.repo.GetArtistBio(artistID); errors.Is(err, repository.ErrNotFound) || errors.Is(err, repository.ErrInvalidType) {
			bio, err = lastFMClient.GetArtistBio(ctx, artist.Name)
			if err != nil {
				sentry.CaptureException(err)
			}
			s.repo.InsertArtistBio(&bio, artistID)
		} else {
			bio = *abio
		}
	}()

	u, err := s.currentUser(ctx, spotifyClient, token)
	if err != nil {
		return types.ArtistInfo{}, errors.Wrap(err, "failed to get user info")
	}

	topTracks, err := spotifyClient.GetArtistsTopTracks(ctx, spotify.ID(artistID), u.Country)
	if err != nil {
		return types.ArtistInfo{}, errors.Wrap(err, "failed to get spotify artist top tracks")
	}

	f := make(map[string]spotify.AudioFeatures)
	var errF error
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer sentry.RecoverWithContext(ctx)
		f, errF = tracksFeatures(ctx, spotifyClient, allTrackID(topTracks))
	}()

	wg.Wait()

	if errF != nil {
		return types.ArtistInfo{}, errors.Wrap(err, "failed to get audio features")
	}

	return types.ArtistInfo{
		Artirst:       *artist,
		TopTracks:     topTracks,
		AudioFeatures: f,
		Bio:           strings.Split(bio.Artist.Bio.Content, "\n"),
	}, nil
}

func (s *Spotify) currentUser(ctx context.Context, client *spotify.Client, token string) (*spotify.PrivateUser, error) {
	u, err := s.repo.GetUser(token)
	if errors.Is(err, repository.ErrNotFound) || errors.Is(err, repository.ErrInvalidType) {
		user, err := client.CurrentUser(ctx)
		s.repo.InsertUser(token, user, nil) // not the end of th world if repo fail to insert
		return user, err
	}
	return u, nil
}
