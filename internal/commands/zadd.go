package commands

import "github.com/hardikphalet/go-redis/internal/store"

type ZAddCommand struct {
	key    string
	score  float64
	member string
}

func (c *ZAddCommand) Execute(store store.Store) (interface{}, error)
