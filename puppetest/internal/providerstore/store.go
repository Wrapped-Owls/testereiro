package providerstore

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
)

type entry struct {
	key      Key
	value    any
	teardown func(context.Context) error
}

type Store struct {
	mu      sync.RWMutex
	entries map[Key]entry
	order   []Key
}

func New() *Store {
	return &Store{
		entries: make(map[Key]entry),
	}
}

func (s *Store) Save(key Key, value any, teardown func(context.Context) error) error {
	if s == nil {
		return errors.New("provider store is nil")
	}
	if key == nil {
		return errors.New("provider key is nil")
	}
	if value == nil {
		return fmt.Errorf("provider %q value is nil", keyLabel(key))
	}

	storeEntry := entry{
		key:      key,
		value:    value,
		teardown: teardown,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entries[key]; !exists {
		s.order = append(s.order, key)
	}
	s.entries[key] = storeEntry
	return nil
}

func SaveProvider[T any](
	s *Store,
	key Key,
	value *T,
	teardown func(context.Context, *T) error,
) error {
	var internalTeardown func(context.Context) error
	if teardown != nil {
		internalTeardown = func(ctx context.Context) error {
			return teardown(ctx, value)
		}
	}

	return s.Save(key, value, internalTeardown)
}

func (s *Store) Load(key Key) (any, bool) {
	if s == nil || key == nil {
		return nil, false
	}

	s.mu.RLock()
	storeEntry, found := s.entries[key]
	s.mu.RUnlock()
	if !found {
		return nil, false
	}

	return storeEntry.value, true
}

func (s *Store) Teardown(ctx context.Context) error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	orderedKeys := slices.Clone(s.order)
	entriesByKey := s.entries
	s.entries = make(map[Key]entry)
	s.order = nil
	s.mu.Unlock()

	var teardownErr error
	for _, key := range slices.Backward(orderedKeys) {
		storeEntry, exists := entriesByKey[key]
		if !exists {
			continue
		}
		if storeEntry.teardown != nil {
			if err := storeEntry.teardown(ctx); err != nil {
				teardownErr = errors.Join(
					teardownErr,
					fmt.Errorf("provider %q: %w", keyLabel(storeEntry.key), err),
				)
			}
		}
	}
	return teardownErr
}

func (s *Store) Keys() []Key {
	if s == nil {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return slices.Clone(s.order)
}

func keyLabel(key Key) string {
	if tagName := key.Tag(); tagName != "" {
		return fmt.Sprintf("%s(%s)", key.Type(), tagName)
	}
	return key.Type().String()
}
