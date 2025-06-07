package commands

import (
	"github.com/hardikphalet/go-redis/internal/commands/options"
	"github.com/hardikphalet/go-redis/internal/store"
)

type ZRangeCommand struct {
	Key     string
	Start   interface{} // Can be int for index-based range or string for score/lex range
	Stop    interface{} // Can be int for index-based range or string for score/lex range
	Options *options.ZRangeOptions
}

func (c *ZRangeCommand) Execute(store store.Store) (interface{}, error) {
	return store.ZRange(c.Key, c.Start, c.Stop, c.Options)
}
