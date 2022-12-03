package types

import "github.com/zmb3/spotify/v2"

type ArtistInfo struct {
	Artirst       spotify.FullArtist               `json:"artist"`
	TopTracks     []spotify.FullTrack              `json:"topTracks"`
	AudioFeatures map[string]spotify.AudioFeatures `json:"trackFeatures"`
	Bio           []string                         `json:"bio"`
}
