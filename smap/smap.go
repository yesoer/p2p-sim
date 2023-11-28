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
	Update(key K, modifier func(value V) V) bool
	Delete(key K)
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

// This function overwrites whatever is stored under "key", be careful with
// overwriting because it could lead to data loss. You probably only want to
// use this for the first write. Else consider Update()
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

// Define a modifier function to update the value under K
func (s *smap[K, V]) Update(key K, modifier func(value V) V) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if v, ok := s.m[key]; ok {
		newV := modifier(v)
		s.m[key] = newV
		return true
	}

	return false
}
