package commands

import "github.com/hardikphalet/go-redis/internal/store"

type GetCommand struct {
	Key string
}

func (c *GetCommand) Execute(store store.Store) (interface{}, error) {
	return store.Get(c.Key)
}
