package commands

import "github.com/hardikphalet/go-redis/internal/store"

type ZRangeCommand struct {
	Key        string
	Start      int
	Stop       int
	WithScores bool
}

func (c *ZRangeCommand) Execute(store store.Store) (interface{}, error) {
	return store.ZRange(c.Key, c.Start, c.Stop, c.WithScores)
}
