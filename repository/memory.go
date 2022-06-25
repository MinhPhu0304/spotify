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
}

type inMemoryRepository struct {
	cache *cache.Cache
}

var ErrNotFound = errors.New("record not found")
var ErrInvalidType = errors.New("invalid struct type store in cache")
var userNamespace = "user-"

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
