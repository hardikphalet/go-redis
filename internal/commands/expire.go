package commands

import (
	"time"

	"github.com/hardikphalet/go-redis/internal/commands/options"
	"github.com/hardikphalet/go-redis/internal/store"
)

type ExpireCommand struct {
	Key     string
	TTL     time.Duration
	Options *options.ExpireOptions
}

func (c *ExpireCommand) Execute(store store.Store) (interface{}, error) {
	return nil, store.Expire(c.Key, c.TTL, c.Options)
}
