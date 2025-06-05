package commands

import "github.com/hardikphalet/go-redis/internal/store"

type Command interface {
	Execute(store store.Store) (interface{}, error)
}
