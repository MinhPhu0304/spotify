package lastfm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/pkg/errors"

	"github.com/MinhPhu0304/spotify/trace"
	"github.com/MinhPhu0304/spotify/types/lastfm"
)

type LastFMClient interface {
	GetArtistBio(ctx context.Context, name string) (lastfm.LastFMBio, error)
}

type lastFMClient struct {
	token  string
	url    string
	client *http.Client
}

var linkRegex = regexp.MustCompile(`<a[^>]*>|<\/a>`)

func Client(token string) LastFMClient {
	c := trace.WrapWithTrace(&http.Client{
		Transport: http.DefaultTransport,
		Timeout:   1 * time.Minute, // Something is really wrong to take 1 min
	})
	return &lastFMClient{
		token:  token,
		url:    "http://ws.audioscrobbler.com/2.0",
		client: c,
	}
}

func (l *lastFMClient) bioURL(artist string) string {
	name := url.QueryEscape(artist)
	return fmt.Sprintf("%s?method=artist.getinfo&artist=%s&api_key=%s&format=json", l.url, name, l.token)
}

func (l *lastFMClient) GetArtistBio(ctx context.Context, name string) (lastfm.LastFMBio, error) {
	url := l.bioURL(name)
	resp, err := l.client.Get(url)
	if err != nil {
		return lastfm.LastFMBio{}, errors.Wrap(err, "failed to send HTTP GET request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return lastfm.LastFMBio{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return lastfm.LastFMBio{}, errors.Wrap(err, "failed to read HTTP response body")
	}

	// Unmarshal the response body into an Artist struct
	artist := lastfm.LastFMBio{}
	err = json.Unmarshal(body, &artist)
	if err != nil {
		return lastfm.LastFMBio{}, errors.Wrap(err, "failed to unmarshal JSON")
	}

	artist.Artist.Bio.Content = linkRegex.ReplaceAllString(artist.Artist.Bio.Content, "")
	artist.Artist.Bio.Summary = linkRegex.ReplaceAllString(artist.Artist.Bio.Summary, "")
	return artist, nil
}
