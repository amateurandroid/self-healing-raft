package store

import "sync"

type KVStore struct {
    mu   sync.RWMutex
    data map[string]string
}

func NewKVStore() *KVStore {
    return &KVStore{data: make(map[string]string)}
}

func (s *KVStore) Put(k, v string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.data[k] = v
}

func (s *KVStore) Get(k string) (string, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    v, ok := s.data[k]
    return v, ok
}
