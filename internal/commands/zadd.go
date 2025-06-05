package commands

import "github.com/hardikphalet/go-redis/internal/store"

// ScoreMember represents a score-member pair for sorted sets
type ScoreMember struct {
	Score  float64
	Member string
}

type ZAddCommand struct {
	Key     string
	Members []ScoreMember
}

func (c *ZAddCommand) Execute(store store.Store) (interface{}, error) {
	added := 0
	for _, sm := range c.Members {
		n, err := store.ZAdd(c.Key, sm.Score, sm.Member)
		if err != nil {
			return nil, err
		}
		added += n
	}
	return added, nil
}
