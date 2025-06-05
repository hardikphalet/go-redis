package commands

import "github.com/hardikphalet/go-redis/internal/store"

type SetCommand struct {
	key   string
	value string
}

func (c *SetCommand) Execute(store store.Store) (interface{}, error)
