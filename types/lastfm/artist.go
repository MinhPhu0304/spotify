package lastfm

type ArtistBio struct {
	Summary string `json:"summary"`
	Content string `json:"content"`
}

type Image struct {
	Link string `json:"#text"`
	Size string `json:"size"`
}

type Tag struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Artist struct {
	Name  string     `json:"name"`
	URL   string     `json:"url"`
	Bio   *ArtistBio `json:"bio"`
	Image []Image
	Tags  struct {
		Tag []Tag `json:"tag"`
	}
}

type LastFMBio struct {
	Artist *Artist `json:"artist"`
}
