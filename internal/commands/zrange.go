package commands

import "github.com/hardikphalet/go-redis/internal/store"

type ZRangeCommand struct {
	key    string
	start  int
	stop   int
	offset int
	count  int
}

func (c *ZRangeCommand) Execute(store store.Store) (interface{}, error)
