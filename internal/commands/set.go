package commands

import "github.com/hardikphalet/go-redis/internal/store"

type SetCommand struct {
	Key   string
	Value string
}

func (c *SetCommand) Execute(store store.Store) (interface{}, error) {
	return nil, store.Set(c.Key, c.Value)
}
