package repository

import (
	"errors"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/zmb3/spotify/v2"
)

type Repository interface {
	GetUser(userToken string) (*spotify.PrivateUser, error)
	InsertUser(userToken string, user *spotify.PrivateUser, duration *time.Duration) error
	GetArtistBio(artistID string) ([]string, error)
	InsertArtistBio(bio []string, artistID string, duration *time.Duration) error
	GetUserTopTracks(userToken string) ([]spotify.FullTrack, error)
	InsertUserTopTracks(topTracks []spotify.FullTrack, userToken string, duration *time.Duration) error
	GetGenres() ([]string, error)
	InsertGenres(genres []string, duration *time.Duration) error
}

type inMemoryRepository struct {
	cache *cache.Cache
}

var ErrNotFound = errors.New("record not found")
var ErrInvalidType = errors.New("invalid struct type store in cache")
var userNamespace = "user-"
var userTopTrackNamespace = "user-top-tracks-"
var artistBioNamespace = "artirst-bio-"
var spotifyGenres = "spotify-genres"

func CreateInMemoryRepo() Repository {
	return &inMemoryRepository{
		cache: cache.New(5*time.Minute, 5*time.Minute),
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

func (r *inMemoryRepository) InsertUserTopTracks(topTracks []spotify.FullTrack, userToken string, duration *time.Duration) error {
	cacheKey := userTopTrackNamespace + userToken
	expr := cache.DefaultExpiration
	if duration != nil {
		expr = *duration
	}
	return r.cache.Add(cacheKey, topTracks, expr)
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

func (r *inMemoryRepository) InsertArtistBio(bio []string, artistID string, duration *time.Duration) error {
	cacheKey := artistBioNamespace + artistID
	expr := cache.DefaultExpiration
	if duration != nil {
		expr = *duration
	}
	return r.cache.Add(cacheKey, bio, expr)
}

func (r *inMemoryRepository) GetArtistBio(artistID string) ([]string, error) {
	cacheKey := userNamespace + artistID
	v, ok := r.cache.Get(cacheKey)
	if !ok {
		return nil, ErrNotFound
	}
	if _, valid := v.([]string); !valid {
		r.cache.Delete(cacheKey)
		return nil, ErrInvalidType
	}
	return v.([]string), nil
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
