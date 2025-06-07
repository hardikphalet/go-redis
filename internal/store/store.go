package store

import (
	"time"

	"github.com/hardikphalet/go-redis/internal/commands/options"
	"github.com/hardikphalet/go-redis/internal/types"
)

// Store defines the interface for the Redis data store
type Store interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, opts *options.SetOptions) (interface{}, error)
	Del(key string) error
	Expire(key string, ttl time.Duration, opts *options.ExpireOptions) error
	TTL(key string) (int, error)
	Keys(pattern string) ([]string, error)

	// Sorted Set operations
	ZAdd(key string, members []types.ScoreMember, opts *options.ZAddOptions) (interface{}, error)
	ZRange(key string, start, stop interface{}, opts *options.ZRangeOptions) ([]interface{}, error)
}
