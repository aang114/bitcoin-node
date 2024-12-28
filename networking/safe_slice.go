package networking

import (
	"sync"
)

type SafeSlice[T any] struct {
	mu sync.RWMutex
	s  []T
}

func NewSafeSlice[T any](size int) *SafeSlice[T] {
	return &SafeSlice[T]{
		s: make([]T, size),
	}
}

func (s *SafeSlice[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.s)
}

func (s *SafeSlice[T]) Append(v T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.s = append(s.s, v)
}

func (s *SafeSlice[T]) AppendMany(vs ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s = append(s.s, vs...)
}

func (s *SafeSlice[T]) Set(slice []T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s = slice
}

func (s *SafeSlice[T]) Get(index int) T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.s[index]
}

func (s *SafeSlice[T]) GetAll() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.s
}

func (s *SafeSlice[T]) Remove(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.s = append(s.s[:index], s.s[index+1:]...)
}

func (s *SafeSlice[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s = s.s[:0]
}
