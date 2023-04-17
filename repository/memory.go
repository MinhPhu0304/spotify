package repository

import (
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/zmb3/spotify/v2"

	"github.com/MinhPhu0304/spotify/types"
	"github.com/MinhPhu0304/spotify/types/lastfm"
)

type Repository interface {
	GetUser(userToken string) (*spotify.PrivateUser, error)
	InsertUser(userToken string, user *spotify.PrivateUser, duration *time.Duration) error
	GetArtistBio(artistID string) (*lastfm.LastFMBio, error)
	InsertArtistBio(bio *lastfm.LastFMBio, artistID string) error
	GetUserTopTracks(userToken string) ([]spotify.FullTrack, error)
	InsertUserTopTracks(topTracks []spotify.FullTrack, userToken string) error
	GetGenres() ([]string, error)
	InsertGenres(genres []string, duration *time.Duration) error
	GetSpotifyClient(token string) (*spotify.Client, error)
	InsertSpotifyClient(token string, client *spotify.Client) error
	GetSpotifyArtist(artistID string) (*spotify.FullArtist, error)
	InsertSpotifyArtist(artist *spotify.FullArtist) error
	GetArtistInfo(artistID string) (*types.ArtistInfo, error)
	InsertArtistInfo(artist *types.ArtistInfo) error
	GetSpotifyFullTrack(trackID string) (*spotify.FullTrack, error)
	InsertSpotifyFullTrack(fullTrack *spotify.FullTrack) error
	GetSong(trackID string) (*types.Song, error)
	InsertSong(song *types.Song) error
	GetTopArtists(userToken string) ([]spotify.FullArtist, error)
	InsertTopArtist(userToken string, artists []spotify.FullArtist) error
}

type inMemoryRepository struct {
	cache *cache.Cache
}

var (
	ErrNotFound               = errors.New("record not found")
	ErrInvalidType            = errors.New("invalid struct type store in cache")
	userNamespace             = "user-"
	artistNamespace           = "artist-"
	artistBioNamespace        = "artirst-bio-"
	topArtistNamespace        = "top-artist-"
	userTopTrackNamespace     = "user-top-tracks-"
	songNamespace             = "song-bio-"
	spotifyClientNamespace    = "spotify-client-"
	spotifyFullTrackNamespace = "spotify-fulltrack-"
	spotifyArtistNamespace    = "spotify-artist-"
	spotifyGenres             = "spotify-genres"
)

func CreateInMemoryRepo() Repository {
	return &inMemoryRepository{
		cache: cache.New(15*time.Minute, 15*time.Minute),
	}
}

func (r *inMemoryRepository) GetUser(userToken string) (*spotify.PrivateUser, error) {
	cacheKey := userNamespace + userToken
	v, ok := r.cache.Get(cacheKey)
	if !ok {
		return nil, ErrNotFound
	}
	if _, valid := v.(spotify.User); !valid {
		r.cache.Delete(cacheKey)
		return nil, ErrInvalidType
	}
	return v.(*spotify.PrivateUser), nil
}

func (r *inMemoryRepository) InsertUser(userToken string, user *spotify.PrivateUser, duration *time.Duration) error {
	expr := cache.DefaultExpiration
	if duration != nil {
		expr = *duration
	}

	return r.cache.Add(userNamespace+userToken, user, expr)
}

func (r *inMemoryRepository) InsertUserTopTracks(topTracks []spotify.FullTrack, userToken string) error {
	cacheKey := userTopTrackNamespace + userToken
	return r.cache.Add(cacheKey, topTracks, cache.DefaultExpiration)
}

func (r *inMemoryRepository) GetUserTopTracks(userToken string) ([]spotify.FullTrack, error) {
	cacheKey := userTopTrackNamespace + userToken
	v, ok := r.cache.Get(cacheKey)
	if !ok {
		return nil, ErrNotFound
	}
	if _, valid := v.([]spotify.FullTrack); !valid {
		r.cache.Delete(cacheKey)
		return nil, ErrInvalidType
	}
	return v.([]spotify.FullTrack), nil
}

func (r *inMemoryRepository) InsertArtistBio(bio *lastfm.LastFMBio, artistID string) error {
	cacheKey := artistBioNamespace + artistID
	return r.cache.Add(cacheKey, bio, 0)
}

func (r *inMemoryRepository) GetArtistBio(artistID string) (*lastfm.LastFMBio, error) {
	cacheKey := artistBioNamespace + artistID
	v, ok := r.cache.Get(cacheKey)
	if !ok {
		return nil, ErrNotFound
	}
	if _, valid := v.(*lastfm.LastFMBio); !valid {
		r.cache.Delete(cacheKey)
		return nil, ErrInvalidType
	}
	return v.(*lastfm.LastFMBio), nil
}

func (r *inMemoryRepository) InsertGenres(genres []string, duration *time.Duration) error {
	expr := cache.DefaultExpiration
	if duration != nil {
		expr = *duration
	}
	return r.cache.Add(spotifyGenres, genres, expr)
}

func (r *inMemoryRepository) GetGenres() ([]string, error) {
	v, ok := r.cache.Get(spotifyGenres)
	if !ok {
		return nil, ErrNotFound
	}
	if _, valid := v.([]string); !valid {
		r.cache.Delete(spotifyGenres)
		return nil, ErrInvalidType
	}
	return v.([]string), nil
}

func (r *inMemoryRepository) GetSpotifyClient(token string) (*spotify.Client, error) {
	cacheKey := spotifyClientNamespace + token
	c, ok := r.cache.Get(cacheKey)
	if !ok {
		return nil, ErrNotFound
	}

	if c, valid := c.(*spotify.Client); !valid {
		r.cache.Delete(cacheKey)
		return nil, ErrInvalidType
	} else {
		return c, nil
	}
}

func (r *inMemoryRepository) InsertSpotifyClient(token string, client *spotify.Client) error {
	cacheKey := spotifyClientNamespace + token
	return r.cache.Add(cacheKey, client, cache.DefaultExpiration)
}

func (r *inMemoryRepository) GetSpotifyArtist(artistID string) (*spotify.FullArtist, error) {
	cacheKey := spotifyArtistNamespace + artistID
	v, ok := r.cache.Get(cacheKey)
	if !ok {
		return nil, ErrNotFound
	}
	if v, valid := v.(*spotify.FullArtist); !valid {
		r.cache.Delete(cacheKey)
		return nil, ErrInvalidType
	} else {
		return v, nil
	}
}

func (r *inMemoryRepository) InsertSpotifyArtist(artist *spotify.FullArtist) error {
	cacheKey := spotifyArtistNamespace + artist.ID.String()
	return r.cache.Add(cacheKey, artist, 2*time.Hour)
}

func (r *inMemoryRepository) GetArtistInfo(artistID string) (*types.ArtistInfo, error) {
	cacheKey := artistNamespace + artistID
	v, ok := r.cache.Get(cacheKey)
	if !ok {
		return nil, ErrNotFound
	}
	if v, valid := v.(*types.ArtistInfo); !valid {
		r.cache.Delete(cacheKey)
		return nil, ErrInvalidType
	} else {
		return v, nil
	}
}

func (r *inMemoryRepository) InsertArtistInfo(artist *types.ArtistInfo) error {
	cacheKey := artistNamespace + artist.Artirst.ID.String()
	return r.cache.Add(cacheKey, artist, 5*time.Minute)
}

func (r *inMemoryRepository) GetSpotifyFullTrack(trackID string) (*spotify.FullTrack, error) {
	cacheKey := spotifyFullTrackNamespace + trackID
	v, ok := r.cache.Get(cacheKey)
	if !ok {
		return nil, ErrNotFound
	}
	if v, valid := v.(*spotify.FullTrack); !valid {
		r.cache.Delete(cacheKey)
		return nil, ErrInvalidType
	} else {
		return v, nil
	}
}

func (r *inMemoryRepository) InsertSpotifyFullTrack(fullTrack *spotify.FullTrack) error {
	cacheKey := spotifyFullTrackNamespace + string(fullTrack.ID)
	return r.cache.Add(cacheKey, fullTrack, 0)
}

func (r *inMemoryRepository) GetSong(trackID string) (*types.Song, error) {
	cacheKey := songNamespace + trackID
	v, ok := r.cache.Get(cacheKey)
	if !ok {
		return nil, ErrNotFound
	}
	if v, valid := v.(*types.Song); !valid {
		r.cache.Delete(cacheKey)
		return nil, ErrInvalidType
	} else {
		return v, nil
	}
}

func (r *inMemoryRepository) InsertSong(song *types.Song) error {
	cacheKey := songNamespace + string(song.Detail.ID)
	return r.cache.Add(cacheKey, song, 10*time.Minute)
}

func (r *inMemoryRepository) GetTopArtists(userToken string) ([]spotify.FullArtist, error) {
	cacheKey := topArtistNamespace + userToken
	v, ok := r.cache.Get(cacheKey)
	if !ok {
		return nil, ErrNotFound
	}
	if v, valid := v.([]spotify.FullArtist); !valid {
		r.cache.Delete(cacheKey)
		return nil, ErrInvalidType
	} else {
		return v, nil
	}
}

func (r *inMemoryRepository) InsertTopArtist(userToken string, artists []spotify.FullArtist) error {
	cacheKey := topArtistNamespace + userToken
	return r.cache.Add(cacheKey, artists, 10*time.Minute)
}
