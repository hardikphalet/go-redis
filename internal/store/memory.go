package store

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/hardikphalet/go-redis/internal/commands/options"
)

// SortedSetMember represents a member in a sorted set
type SortedSetMember struct {
	Score  float64
	Member string
}

// SortedSet represents a Redis sorted set
type SortedSet struct {
	dict map[string]float64 // For O(1) member lookups
	sl   *skiplist          // For ordered operations
}

// Add adds or updates a member in the sorted set
func (s *SortedSet) Add(score float64, member string) bool {
	if s.dict == nil {
		s.dict = make(map[string]float64)
		s.sl = newSkiplist()
	}

	// Check if member exists
	oldScore, exists := s.dict[member]
	if exists && oldScore == score {
		return false
	}

	// Update or add to map
	s.dict[member] = score

	// Update or add to skiplist
	return s.sl.insert(score, member)
}

// Range returns a range of members from the sorted set
func (s *SortedSet) Range(start, stop int, withScores bool) []interface{} {
	if s == nil || s.sl == nil || len(s.dict) == 0 {
		return []interface{}{}
	}

	// Get range from skiplist
	nodes := s.sl.getRange(start, stop)

	// Prepare result
	result := make([]interface{}, 0, len(nodes)*2)
	for _, node := range nodes {
		result = append(result, node.member)
		if withScores {
			result = append(result, node.score)
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

// Between reading for expiry and reading from the map, there is a race condition
// func (s *MemoryStore) Get(key string) (interface{}, error) {
// 	s.mu.RLock()
// 	expired := s.isExpired(key)
// 	s.mu.RUnlock()

// 	if expired {
// 		s.Del(key)
// 		return nil, nil
// 	}

//		s.mu.RLock()
//		defer s.mu.RUnlock()
//		if val, ok := s.data[key]; ok {
//			return val, nil
//		}
//		return nil, nil
//	}
func (s *MemoryStore) Get(key string) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isExpired(key) {
		delete(s.data, key)
		delete(s.expires, key)
		return nil, nil
	}

	if val, ok := s.data[key]; ok {
		return val, nil
	}

	return nil, nil
}

func (s *MemoryStore) Set(key string, value interface{}, opts *options.SetOptions) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if key exists
	exists := false
	var oldValue interface{}
	if val, ok := s.data[key]; ok {
		exists = true
		oldValue = val
	}

	// Handle NX option - only set if key doesn't exist
	if opts != nil && opts.IsNX() && exists {
		return nil, fmt.Errorf("key already exists")
	}

	// Handle XX option - only set if key exists
	if opts != nil && opts.IsXX() && !exists {
		return nil, fmt.Errorf("key does not exist")
	}

	// Store the value
	s.data[key] = value

	// Handle expiry
	if opts != nil {
		if opts.IsKEEPTTL() {
			// Keep existing TTL if any
			if _, ok := s.expires[key]; !ok {
				delete(s.expires, key)
			}
		} else if opts.ExpiryType != "" {
			// Set new expiry
			s.expires[key] = opts.ExpiryTime
		} else {
			// No expiry specified, remove any existing expiry
			delete(s.expires, key)
		}
	} else {
		// No options specified, remove any existing expiry
		delete(s.expires, key)
	}

	// Return old value if GET option is set
	if opts != nil && opts.IsGET() {
		return oldValue, nil
	}

	return nil, nil
}

func (s *MemoryStore) Del(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	delete(s.expires, key)
	return nil
}

func (s *MemoryStore) Expire(key string, ttl time.Duration, opts *options.ExpireOptions) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.data[key]; !exists {
		return fmt.Errorf("key does not exist")
	}

	// Handle options
	if opts != nil {
		// Check if key has an existing expiry
		hasExpiry := false
		if expiry, ok := s.expires[key]; ok {
			hasExpiry = !time.Now().After(expiry)
		}

		// Handle NX option - only set expiry if key has no expiry
		if opts.IsNX() && hasExpiry {
			return fmt.Errorf("key already has an expiry")
		}

		// Handle XX option - only set expiry if key has an existing expiry
		if opts.IsXX() && !hasExpiry {
			return fmt.Errorf("key has no expiry")
		}

		// Handle GT option - only set expiry if new expiry is greater than current one
		if opts.IsGT() && hasExpiry {
			currentTTL := time.Until(s.expires[key])
			if ttl <= currentTTL {
				return fmt.Errorf("new expiry is not greater than current one")
			}
		}

		// Handle LT option - only set expiry if new expiry is less than current one
		if opts.IsLT() && hasExpiry {
			currentTTL := time.Until(s.expires[key])
			if ttl >= currentTTL {
				return fmt.Errorf("new expiry is not less than current one")
			}
		}
	}

	if ttl <= 0 {
		delete(s.expires, key)
		delete(s.data, key)
		return nil
	}

	s.expires[key] = time.Now().Add(ttl)
	return nil
}

func (s *MemoryStore) TTL(key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if expiry, ok := s.expires[key]; ok {
		if ttl := time.Until(expiry); ttl > 0 {
			return int(ttl.Seconds()), nil
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
