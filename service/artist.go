package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/MinhPhu0304/spotify/repository"
	"github.com/MinhPhu0304/spotify/trace"
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

	client := spotify.New(trace.WrapWithTrace(s.spotifyAuth.Client(ctx, &oauth2.Token{AccessToken: token})))
	artist, err := client.GetArtist(ctx, spotify.ID(artistID))

	if err != nil {
		return ArtistInfo{}, errors.Wrap(err, "failed to get spotify artist")
	}

	u, err := s.getCurrentUser(ctx, client, token)
	if err != nil {
		return ArtistInfo{}, errors.Wrap(err, "failed to get user info")
	}

	topTracks, err := client.GetArtistsTopTracks(ctx, spotify.ID(artistID), u.Country)
	if err != nil {
		return ArtistInfo{}, errors.Wrap(err, "failed to get spotify artist top tracks")
	}

	bio := make([]string, 0)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		bio, err = s.getArtistBio(ctx, artist.Name)
		if err != nil {
			sentry.CaptureException(err)
		}
		wg.Done()
	}()

	f := make(map[string]spotify.AudioFeatures)
	var errF error
	wg.Add(1)
	go func() {
		f, errF = tracksFeatures(ctx, client, allTrackID(topTracks))
		wg.Done()
	}()

	wg.Wait()

	if errF != nil {
		return ArtistInfo{}, errors.Wrap(err, "failed to get audio features")
	}

	return ArtistInfo{
		Artirst:       *artist,
		TopTracks:     topTracks,
		AudioFeatures: f,
		Bio:           bio,
	}, nil
}

func (s *Service) getArtistBio(ctx context.Context, name string) ([]string, error) {
	c := trace.DefaultTracedClient()
	r, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://www.last.fm/music/%s/+wiki", name), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(r)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("status code error: %d %s", resp.StatusCode, resp.Status))
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse body")
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

func (s *Service) getCurrentUser(ctx context.Context, client *spotify.Client, token string) (*spotify.PrivateUser, error) {
	u, err := s.repo.GetUser(token)
	if errors.Is(err, repository.ErrNotFound) || errors.Is(err, repository.ErrInvalidType) {
		user, err := client.CurrentUser(ctx)
		s.repo.InsertUser(token, user, nil) // not the end of th world if repo fail to insert
		return user, err
	}
	return u, nil
}
