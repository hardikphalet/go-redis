package store

import (
	"fmt"
	"sync"
	"time"
)

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

	// For now, we'll just return all keys
	// TODO: Implement pattern matching
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		if !s.isExpired(k) {
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
