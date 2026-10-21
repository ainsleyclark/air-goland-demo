package store

import (
	"context"
	"sync"
)

// Memory is an in-memory Store. It is safe for concurrent use.
type Memory struct {
	mu     sync.Mutex
	events map[string]Event
}

var _ Store = (*Memory)(nil)

// NewMemory returns an empty Memory store.
func NewMemory() *Memory {
	return &Memory{events: make(map[string]Event)}
}

func (m *Memory) Get(_ context.Context, id string) (Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ev, ok := m.events[id]
	if !ok {
		return Event{}, ErrNotFound
	}
	return ev, nil
}

func (m *Memory) Save(_ context.Context, ev Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.events[ev.ID]; ok {
		return ErrDuplicate
	}
	m.events[ev.ID] = ev
	return nil
}
