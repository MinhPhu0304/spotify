package spotify

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/MinhPhu0304/spotify/repository"
	"github.com/MinhPhu0304/spotify/trace"
	"github.com/MinhPhu0304/spotify/types"
	"github.com/PuerkitoBio/goquery"
	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify/v2"
)

func (s *Spotify) Artist(ctx context.Context, token string, artistID string) (types.ArtistInfo, error) {
	spotifyClient := s.clientWithTrace(ctx, token)
	artist, err := spotifyClient.GetArtist(ctx, spotify.ID(artistID))

	if err != nil {
		return types.ArtistInfo{}, errors.Wrap(err, "failed to get spotify artist")
	}

	u, err := s.currentUser(ctx, spotifyClient, token)
	if err != nil {
		return types.ArtistInfo{}, errors.Wrap(err, "failed to get user info")
	}

	topTracks, err := spotifyClient.GetArtistsTopTracks(ctx, spotify.ID(artistID), u.Country)
	if err != nil {
		return types.ArtistInfo{}, errors.Wrap(err, "failed to get spotify artist top tracks")
	}

	bio := make([]string, 0)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer sentry.RecoverWithContext(ctx)
		bio, err = s.getArtistBio(ctx, artist.Name, artistID)
		if err != nil {
			sentry.CaptureException(err)
		}
	}()

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
		Bio:           bio,
	}, nil
}

func (s *Spotify) getArtistBio(ctx context.Context, name string, artistID string) ([]string, error) {
	if bio, err := s.repo.GetArtistBio(artistID); err != nil && len(bio) != 0 {
		return bio, nil
	}

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
	defer func() {
		s.repo.InsertArtistBio(bio, artistID, nil)
	}()
	if noWiki {
		return bio, nil
	}

	doc.Find("div.wiki-content p").Each(func(_ int, s *goquery.Selection) {
		bio = append(bio, s.Text())
	})
	return bio, nil
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
