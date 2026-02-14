package providerstore

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"weak"
)

type cell struct {
	value any
}

type entry struct {
	key       Key
	weakValue weak.Pointer[cell]
	keepAlive func()
	teardown  func(context.Context) error
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

	resourceCell := &cell{value: value}
	storeEntry := entry{
		key:       key,
		weakValue: weak.Make(resourceCell),
		keepAlive: func() {
			runtime.KeepAlive(resourceCell)
			runtime.KeepAlive(value)
		},
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

	resourceCell := storeEntry.weakValue.Value()
	if resourceCell == nil {
		return nil, false
	}
	return resourceCell.value, true
}

func (s *Store) Teardown(ctx context.Context) error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	orderedEntries := make([]entry, 0, len(s.order))
	for _, key := range s.order {
		orderedEntries = append(orderedEntries, s.entries[key])
	}
	s.entries = make(map[Key]entry)
	s.order = nil
	s.mu.Unlock()

	var teardownErr error
	for idx := len(orderedEntries) - 1; idx >= 0; idx-- {
		storeEntry := orderedEntries[idx]
		if storeEntry.teardown != nil {
			if err := storeEntry.teardown(ctx); err != nil {
				teardownErr = errors.Join(
					teardownErr,
					fmt.Errorf("provider %q: %w", keyLabel(storeEntry.key), err),
				)
			}
		}
		if storeEntry.keepAlive != nil {
			storeEntry.keepAlive()
		}
	}
	return teardownErr
}

func keyLabel(key Key) string {
	if tagName := key.Tag(); tagName != "" {
		return fmt.Sprintf("%s(%s)", key.Type(), tagName)
	}
	return key.Type().String()
}
