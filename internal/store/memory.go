package store

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/hardikphalet/go-redis/internal/commands/options"
	"github.com/hardikphalet/go-redis/internal/types"
)

// SortedSetMember represents a member in a sorted set
type SortedSetMember struct {
	Score  float64
	Member string
}

// SortedSet represents a Redis sorted set
type SortedSet struct {
	dict    map[string]float64 // For O(1) member lookups
	sl      *skiplist          // For ordered operations
	scores  []float64
	members []string
}

// Add adds or updates a member in the sorted set
func (s *SortedSet) Add(member string, score float64) {
	// Update dictionary
	s.dict[member] = score

	// Update skiplist
	s.sl.insert(score, member)

	// Update slices
	s.scores = append(s.scores, score)
	s.members = append(s.members, member)
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

// RangeByScore returns elements with scores between min and max
func (s *SortedSet) RangeByScore(min, max float64, rev bool) []interface{} {
	var result []interface{}

	if rev {
		// Reverse order
		for i := len(s.scores) - 1; i >= 0; i-- {
			score := s.scores[i]
			if score >= min && score <= max {
				member := s.members[i]
				result = append(result, member)
			}
		}
	} else {
		// Forward order
		for i := 0; i < len(s.scores); i++ {
			score := s.scores[i]
			if score >= min && score <= max {
				member := s.members[i]
				result = append(result, member)
			}
		}
	}

	return result
}

// RangeByLex returns elements with lexicographical ordering between min and max
func (s *SortedSet) RangeByLex(min, max string, rev bool) []interface{} {
	var result []interface{}

	if rev {
		// Reverse order
		for i := len(s.members) - 1; i >= 0; i-- {
			member := s.members[i]
			if member >= min && member <= max {
				result = append(result, member)
			}
		}
	} else {
		// Forward order
		for i := 0; i < len(s.members); i++ {
			member := s.members[i]
			if member >= min && member <= max {
				result = append(result, member)
			}
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

func (s *MemoryStore) ZAdd(key string, members []types.ScoreMember, opts *options.ZAddOptions) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if key exists and is a sorted set
	var zset *SortedSet
	if val, exists := s.data[key]; exists {
		var ok bool
		zset, ok = val.(*SortedSet)
		if !ok {
			return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
	} else {
		zset = &SortedSet{}
		s.data[key] = zset
	}

	// Handle INCR option - only one score-member pair allowed
	if opts != nil && opts.IsINCR() {
		if len(members) != 1 {
			return nil, fmt.Errorf("INCR option requires exactly one score-member pair")
		}
		sm := members[0]
		oldScore, exists := zset.dict[sm.Member]
		if !exists {
			return nil, fmt.Errorf("element does not exist")
		}
		newScore := oldScore + sm.Score
		zset.Add(sm.Member, newScore)
		return newScore, nil
	}

	// Handle other options
	changed := 0
	for _, sm := range members {
		oldScore, exists := zset.dict[sm.Member]

		// Handle NX option - only add new elements
		if opts != nil && opts.IsNX() && exists {
			continue
		}

		// Handle XX option - only update existing elements
		if opts != nil && opts.IsXX() && !exists {
			continue
		}

		// Handle GT option - only update if new score is greater
		if opts != nil && opts.IsGT() && exists && sm.Score <= oldScore {
			continue
		}

		// Handle LT option - only update if new score is less
		if opts != nil && opts.IsLT() && exists && sm.Score >= oldScore {
			continue
		}

		// Add or update the element
		zset.Add(sm.Member, sm.Score)
		changed++
	}

	// Return number of changed elements if CH option is set
	if opts != nil && opts.IsCH() {
		return changed, nil
	}

	// Return number of new elements added
	return len(members), nil
}

func (s *MemoryStore) ZRange(key string, start, stop interface{}, opts *options.ZRangeOptions) ([]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if key exists and is a sorted set
	if val, exists := s.data[key]; exists {
		if zset, ok := val.(*SortedSet); ok {
			var result []interface{}

			// Handle different range types
			if opts != nil && opts.IsByScore() {
				// Convert start and stop to float64 for score-based range
				minScore, ok := start.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid score range start")
				}
				maxScore, ok := stop.(float64)
				if !ok {
					return nil, fmt.Errorf("invalid score range stop")
				}
				result = zset.RangeByScore(minScore, maxScore, opts.IsRev())
			} else if opts != nil && opts.IsByLex() {
				// Convert start and stop to string for lexicographical range
				minLex, ok := start.(string)
				if !ok {
					return nil, fmt.Errorf("invalid lex range start")
				}
				maxLex, ok := stop.(string)
				if !ok {
					return nil, fmt.Errorf("invalid lex range stop")
				}
				result = zset.RangeByLex(minLex, maxLex, opts.IsRev())
			} else {
				// Convert start and stop to int for index-based range
				startIdx, ok := start.(int)
				if !ok {
					return nil, fmt.Errorf("invalid index range start")
				}
				stopIdx, ok := stop.(int)
				if !ok {
					return nil, fmt.Errorf("invalid index range stop")
				}
				result = zset.Range(startIdx, stopIdx, opts != nil && opts.IsWithScores())
			}

			// Apply LIMIT if specified
			if opts != nil && opts.Limit.Count > 0 {
				offset := opts.Limit.Offset
				count := opts.Limit.Count
				if offset >= len(result) {
					return []interface{}{}, nil
				}
				end := offset + count
				if end > len(result) {
					end = len(result)
				}
				result = result[offset:end]
			}

			return result, nil
		}
		return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	return []interface{}{}, nil // Empty array for non-existent key
}
