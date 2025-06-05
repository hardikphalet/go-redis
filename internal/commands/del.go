package commands

import "github.com/hardikphalet/go-redis/internal/store"

type DelCommand struct {
	keys []string
}

func (c *DelCommand) Execute(store store.Store) (interface{}, error)
