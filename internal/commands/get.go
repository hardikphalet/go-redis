package commands

import "github.com/hardikphalet/go-redis/internal/store"

type GetCommand struct {
	key string
}

func (c *GetCommand) Execute(store store.Store) (interface{}, error)
