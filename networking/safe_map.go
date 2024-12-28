package networking

import (
	"sync"
)

type SafeMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		m: make(map[K]V),
	}
}

func (s *SafeMap[K, V]) Get(k K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.m[k]

	return v, ok
}

func (s *SafeMap[K, V]) Set(k K, v V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[k] = v
}

func (s *SafeMap[K, V]) Delete(k K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, k)
}

func (s *SafeMap[K, V]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.m)
}

func (s *SafeMap[K, V]) Pop() (K, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for item := range s.m {
		delete(s.m, item)
		return item, true
	}
	return *new(K), false
}

func (s *SafeMap[K, V]) GetRandomKey() (K, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for item := range s.m {
		return item, true
	}
	return *new(K), false
}

func (s *SafeMap[K, V]) Keys() []K {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]K, 0, len(s.m))
	for k := range s.m {
		keys = append(keys, k)
	}
	return keys
}

func (s *SafeMap[K, V]) Values() []V {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := make([]V, 0, len(s.m))
	for _, v := range s.m {
		values = append(values, v)
	}
	return values
}

func (s *SafeMap[K, V]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m = make(map[K]V)
}
