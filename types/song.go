package types

import "github.com/zmb3/spotify/v2"

type Song struct {
	Detail          spotify.FullTrack     `json:"detail"`
	Features        spotify.AudioFeatures `json:"features"`
	Recommendations []spotify.SimpleTrack `json:"recommendations"`
}
