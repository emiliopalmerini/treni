package cache

import (
	"context"
	"sync"
	"time"
)

type entry struct {
	value     []byte
	expiresAt time.Time
}

// Memory is an in-memory cache implementation.
type Memory struct {
	mu      sync.RWMutex
	entries map[string]*entry
}

// NewMemory creates a new in-memory cache.
func NewMemory() *Memory {
	return &Memory{
		entries: make(map[string]*entry),
	}
}

func (m *Memory) Get(_ context.Context, key string) ([]byte, bool) {
	m.mu.RLock()
	e, ok := m.entries[key]
	m.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if time.Now().After(e.expiresAt) {
		m.mu.Lock()
		delete(m.entries, key)
		m.mu.Unlock()
		return nil, false
	}

	return e.value, true
}

func (m *Memory) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	m.entries[key] = &entry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	m.mu.Unlock()
	return nil
}

func (m *Memory) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	delete(m.entries, key)
	m.mu.Unlock()
	return nil
}

func (m *Memory) Clear(_ context.Context) error {
	m.mu.Lock()
	m.entries = make(map[string]*entry)
	m.mu.Unlock()
	return nil
}
