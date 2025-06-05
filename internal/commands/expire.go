package commands

import (
	"time"

	"github.com/hardikphalet/go-redis/internal/store"
)

type ExpireCommand struct {
	key string
	ttl time.Duration
}

func (c *ExpireCommand) Execute(store store.Store) (interface{}, error)
