package store

import (
	"fmt"
	"regexp"
	"sort"
	"sync"
	"time"
)

// SortedSetMember represents a member in a sorted set
type SortedSetMember struct {
	Score  float64
	Member string
}

// SortedSet represents a Redis sorted set
type SortedSet struct {
	Members []SortedSetMember
}

// Add adds or updates a member in the sorted set
func (s *SortedSet) Add(score float64, member string) bool {
	// Check if member exists
	for i, m := range s.Members {
		if m.Member == member {
			if m.Score != score {
				// Update score
				s.Members[i].Score = score
				// Re-sort the set
				sort.Slice(s.Members, func(i, j int) bool {
					if s.Members[i].Score == s.Members[j].Score {
						return s.Members[i].Member < s.Members[j].Member
					}
					return s.Members[i].Score < s.Members[j].Score
				})
				return false
			}
			return false
		}
	}

	// Add new member
	s.Members = append(s.Members, SortedSetMember{Score: score, Member: member})
	// Sort the set
	sort.Slice(s.Members, func(i, j int) bool {
		if s.Members[i].Score == s.Members[j].Score {
			return s.Members[i].Member < s.Members[j].Member
		}
		return s.Members[i].Score < s.Members[j].Score
	})
	return true
}

// Range returns a range of members from the sorted set
func (s *SortedSet) Range(start, stop int, withScores bool) []interface{} {
	if s == nil || len(s.Members) == 0 {
		return []interface{}{}
	}

	// Handle negative indices
	if start < 0 {
		start = len(s.Members) + start
	}
	if stop < 0 {
		stop = len(s.Members) + stop
	}

	// Boundary checks
	if start < 0 {
		start = 0
	}
	if stop >= len(s.Members) {
		stop = len(s.Members) - 1
	}
	if start > stop {
		return []interface{}{}
	}

	// Prepare result
	result := make([]interface{}, 0)
	for i := start; i <= stop; i++ {
		result = append(result, s.Members[i].Member)
		if withScores {
			result = append(result, s.Members[i].Score)
		}
	}

	return result
}

type MemoryStore struct {
	data    map[string]interface{}
	expires map[string]time.Time
	mu      sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data:    make(map[string]interface{}),
		expires: make(map[string]time.Time),
	}
}

func (s *MemoryStore) Get(key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.isExpired(key) {
		s.mu.RUnlock()
		s.Del(key)
		s.mu.RLock()
		return nil, nil
	}

	if val, ok := s.data[key]; ok {
		return val, nil
	}
	return nil, nil
}

func (s *MemoryStore) Set(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
	delete(s.expires, key) // Reset expiration on SET
	return nil
}

func (s *MemoryStore) Del(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	delete(s.expires, key)
	return nil
}

func (s *MemoryStore) Expire(key string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[key]; !exists {
		return fmt.Errorf("key does not exist")
	}

	if ttl <= 0 {
		delete(s.expires, key)
		delete(s.data, key)
		return nil
	}

	s.expires[key] = time.Now().Add(ttl)
	return nil
}

func (s *MemoryStore) TTL(key string) (time.Duration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if expiry, ok := s.expires[key]; ok {
		if ttl := time.Until(expiry); ttl > 0 {
			return ttl, nil
		}
		return -2, nil // -2 indicates that the key has expired
	}
	if _, ok := s.data[key]; !ok {
		return -2, nil // Key doesn't exist
	}
	return -1, nil // -1 indicates no expiry set
}

func (s *MemoryStore) Keys(pattern string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		if !s.isExpired(k) && matchPattern(k, pattern) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (s *MemoryStore) isExpired(key string) bool {
	if expiry, ok := s.expires[key]; ok {
		return time.Now().After(expiry)
	}
	return false
}

// matchPattern implements Redis-style pattern matching
// Supports:
// * - matches zero or more characters
// ? - matches exactly one character
// [...] - matches any character within the brackets
// [^...] - matches any character not within the brackets
func matchPattern(str, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Convert the pattern to a regex pattern
	regexPattern := ""
	i := 0
	for i < len(pattern) {
		switch pattern[i] {
		case '*':
			regexPattern += ".*"
		case '?':
			regexPattern += "."
		case '[':
			j := i + 1
			if j < len(pattern) && pattern[j] == '^' {
				regexPattern += "[^"
				j++
			} else {
				regexPattern += "["
			}
			for j < len(pattern) && pattern[j] != ']' {
				if pattern[j] == '\\' && j+1 < len(pattern) {
					regexPattern += "\\" + string(pattern[j+1])
					j += 2
				} else {
					regexPattern += string(pattern[j])
					j++
				}
			}
			if j < len(pattern) && pattern[j] == ']' {
				regexPattern += "]"
				i = j
			} else {
				regexPattern += "\\["
			}
		case '\\':
			if i+1 < len(pattern) {
				regexPattern += "\\" + string(pattern[i+1])
				i++
			} else {
				regexPattern += "\\\\"
			}
		default:
			if c := string(pattern[i]); c == "." || c == "+" || c == "(" || c == ")" || c == "|" || c == "{" || c == "}" || c == "$" || c == "^" {
				regexPattern += "\\" + c
			} else {
				regexPattern += c
			}
		}
		i++
	}

	// Use the regex package to match
	matched, err := regexp.MatchString("^"+regexPattern+"$", str)
	if err != nil {
		return false
	}
	return matched
}

func (s *MemoryStore) ZAdd(key string, score float64, member string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if key exists and is a sorted set
	var zset *SortedSet
	if val, exists := s.data[key]; exists {
		var ok bool
		zset, ok = val.(*SortedSet)
		if !ok {
			return 0, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
	} else {
		zset = &SortedSet{}
		s.data[key] = zset
	}

	// Add member to sorted set
	if zset.Add(score, member) {
		return 1, nil // New member added
	}
	return 0, nil // Member updated
}

func (s *MemoryStore) ZRange(key string, start, stop int, withScores bool) ([]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if key exists and is a sorted set
	if val, exists := s.data[key]; exists {
		if zset, ok := val.(*SortedSet); ok {
			return zset.Range(start, stop, withScores), nil
		}
		return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return []interface{}{}, nil // Empty array for non-existent key
}
