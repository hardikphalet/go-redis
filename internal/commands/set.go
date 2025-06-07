package commands

import (
	"github.com/hardikphalet/go-redis/internal/commands/options"
	"github.com/hardikphalet/go-redis/internal/store"
)

type SetCommand struct {
	Key     string
	Value   string
	Options *options.SetOptions
}

func (c *SetCommand) Execute(store store.Store) (interface{}, error) {
	return store.Set(c.Key, c.Value, c.Options)
}
