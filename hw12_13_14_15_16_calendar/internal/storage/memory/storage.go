package memorystorage

import (
	"context"
	"sync"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	mu     sync.RWMutex
	events map[string]storage.Event
}

func New() *Storage {
	return &Storage{
		events: make(map[string]storage.Event),
	}
}

func (s *Storage) AddEvent(_ context.Context, e storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[e.ID]; exists {
		return storage.ErrExists
	}

	s.events[e.ID] = e
	return nil
}

func (s *Storage) UpdateEvent(_ context.Context, e storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[e.ID]; !exists {
		return storage.ErrNotFound
	}

	s.events[e.ID] = e
	return nil
}

func (s *Storage) DeleteEvent(_ context.Context, e storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.events[e.ID]; !exists {
		return storage.ErrNotFound
	}

	delete(s.events, e.ID)
	return nil
}

func (s *Storage) ListEvents(_ context.Context) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]storage.Event, 0, len(s.events))
	for _, e := range s.events {
		events = append(events, e)
	}
	return events, nil
}

func (s *Storage) GetEvent(_ context.Context, id string) (storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, exists := s.events[id]
	if !exists {
		return storage.Event{}, storage.ErrNotFound
	}
	return e, nil
}
