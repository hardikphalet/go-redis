package commands

import "github.com/hardikphalet/go-redis/internal/store"

type DelCommand struct {
	Keys []string
}

func (c *DelCommand) Execute(store store.Store) (interface{}, error) {
	var deleted int
	for _, key := range c.Keys {
		err := store.Del(key)
		if err == nil {
			deleted++
		}
	}
	return deleted, nil
}
