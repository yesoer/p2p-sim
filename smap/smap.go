package smap

import (
	"sync"
)

/*
* A generic map that is safe for concurrent use, but not optimized for anything.
 */

type SMap[K comparable, V any] interface {
	Load(key K) (V, bool)
	Store(key K, value V)
}

type smap[K comparable, V any] struct {
	m  map[K]V
	mu sync.RWMutex
}

func NewSMap[K comparable, V any]() SMap[K, V] {
	m := make(map[K]V)
	return &smap[K, V]{m, sync.RWMutex{}}
}

func (s *smap[K, V]) Load(key K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[key]
	return v, ok
}

func (s *smap[K, V]) Store(key K, value V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}

func (s *smap[K, V]) Delete(key K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key)
}
