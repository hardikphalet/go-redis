package store

import (
	"time"
)

// Store defines the interface for the Redis data store
type Store interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
	Del(key string) error
	Expire(key string, ttl time.Duration) error
	TTL(key string) (time.Duration, error)
	Keys(pattern string) ([]string, error)
}
