package store

import (
	"encoding/json"
	"os"
	"sync"
)

type JSONStore[T any] struct {
	mu   sync.RWMutex
	path string
}

func NewJSONStore[T any](path string) *JSONStore[T] {
	return &JSONStore[T]{path: path}
}

func (s *JSONStore[T]) Load() (*T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			var zero T
			return &zero, nil
		}
		return nil, err
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (s *JSONStore[T]) Save(v *T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *JSONStore[T]) LoadInto(v *T) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, v)
}

func (s *JSONStore[T]) SaveFrom(v *T) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}
