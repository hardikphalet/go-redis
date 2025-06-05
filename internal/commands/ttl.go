package commands

import "github.com/hardikphalet/go-redis/internal/store"

type TtlCommand struct {
	Key string
}

func (c *TtlCommand) Execute(store store.Store) (interface{}, error) {
	ttl, err := store.TTL(c.Key)
	if err != nil {
		return nil, err
	}
	return ttl, nil
}
