package commands

import "github.com/hardikphalet/go-redis/internal/store"

type KeysCommand struct {
	Pattern string
}

func (c *KeysCommand) Execute(store store.Store) (interface{}, error) {
	return store.Keys(c.Pattern)
}
