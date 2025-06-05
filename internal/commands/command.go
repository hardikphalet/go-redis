package commands

import (
	"github.com/hardikphalet/go-redis/internal/store"
)

type Command interface {
	Execute(store store.Store) (interface{}, error)
}

type CommandCommand struct{}

func (c *CommandCommand) Execute(store store.Store) (interface{}, error) {
	return []interface{}{
		"set",
		map[string]interface{}{
			"arity":     3,
			"flags":     []string{"write", "denyoom"},
			"key_start": 1,
			"key_end":   1,
			"key_step":  1,
		},
		"get",
		map[string]interface{}{
			"arity":     2,
			"flags":     []string{"readonly", "fast"},
			"key_start": 1,
			"key_end":   1,
			"key_step":  1,
		},
		// ...
	}, nil
}
