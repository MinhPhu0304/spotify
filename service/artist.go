package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type ArtistInfo struct {
	Artirst       spotify.FullArtist               `json:"artist"`
	TopTracks     []spotify.FullTrack              `json:"topTracks"`
	AudioFeatures map[string]spotify.AudioFeatures `json:"trackFeatures"`
	Bio           []string                         `json:"bio"`
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

	user, err := client.CurrentUser(ctx)
	if err != nil {
		return ArtistInfo{}, errors.Wrap(err, "failed to get user info")
	}

	topTracks, err := client.GetArtistsTopTracks(ctx, spotify.ID(artistID), user.Country)
	if err != nil {
		return ArtistInfo{}, errors.Wrap(err, "failed to get spotify artist top tracks")
	}

	bio, err := s.getArtistBio(ctx, artist.Name)
	if err != nil {
		sentry.CaptureException(err)
	}

	tracksFeatures, err := tracksFeatures(ctx, client, allTrackID(topTracks))
	if err != nil {
		return ArtistInfo{}, errors.Wrap(err, "failed to get audio features")
	}

	return ArtistInfo{
		Artirst:       *artist,
		TopTracks:     topTracks,
		AudioFeatures: tracksFeatures,
		Bio:           bio,
	}, nil
}

func (s *Service) getArtistBio(ctx context.Context, name string) ([]string, error) {
	resp, err := http.Get(fmt.Sprintf("https://www.last.fm/music/%s/+wiki", name))

	if err != nil {
		return []string{}, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []string{}, errors.New(fmt.Sprintf("status code error: %d %s", resp.StatusCode, resp.Status))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return []string{}, errors.Wrap(err, "failed to parse body")
	}

	bio := make([]string, 0)
	noWiki := false
	doc.Find("no-data-message--wiki").Each(func(_ int, s *goquery.Selection) {
		noWiki = true
	})

	if noWiki {
		return bio, nil
	}

	doc.Find("div.wiki-content p").Each(func(_ int, s *goquery.Selection) {
		bio = append(bio, s.Text())
	})

	return bio, nil
}
