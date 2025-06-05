package commands

import "github.com/hardikphalet/go-redis/internal/store"

type TtlCommand struct {
	key string
}

func (c *TtlCommand) Execute(store store.Store) (interface{}, error)
