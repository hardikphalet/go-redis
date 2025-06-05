package commands

import "github.com/hardikphalet/go-redis/internal/store"

type KeysCommand struct {
	pattern string
}

func (c *KeysCommand) Execute(store store.Store) (interface{}, error)
