package memorystorage

import (
	"context"
	"sync"
	"time"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
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

func (s *Storage) AddEvent(ctx context.Context, e storage.Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := e.ID.String()
	if _, exists := s.events[id]; exists {
		return storage.ErrExists
	}

	s.events[id] = e
	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, e storage.Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := e.ID.String()
	if _, exists := s.events[id]; !exists {
		return storage.ErrNotFound
	}

	s.events[id] = e
	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, e storage.Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	id := e.ID.String()
	if _, exists := s.events[id]; !exists {
		return storage.ErrNotFound
	}

	delete(s.events, id)
	return nil
}

func (s *Storage) ListEvents(ctx context.Context) ([]storage.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := make([]storage.Event, 0, len(s.events))
	for _, e := range s.events {
		events = append(events, e)
	}
	return events, nil
}

func (s *Storage) GetEvent(ctx context.Context, uuid uuid.UUID) (storage.Event, error) {
	select {
	case <-ctx.Done():
		return storage.Event{}, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	id := uuid.String()
	e, exists := s.events[id]
	if !exists {
		return storage.Event{}, storage.ErrNotFound
	}
	return e, nil
}

func (s *Storage) GetEventsByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]storage.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []storage.Event
	for _, event := range s.events {
		if !event.DateStart.Before(startTime) && event.DateStart.Before(endTime) {
			result = append(result, event)
		}
	}

	return result, nil
}
