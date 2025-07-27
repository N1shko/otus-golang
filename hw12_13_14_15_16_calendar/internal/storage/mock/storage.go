package mockstorage

import (
	"context"
	"time"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type Storage struct {
	mock.Mock
}

func New() *Storage {
	return &Storage{}
}

func (m *Storage) AddEvent(ctx context.Context, e storage.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *Storage) UpdateEvent(ctx context.Context, e storage.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *Storage) DeleteEvent(ctx context.Context, e storage.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m *Storage) ListEvents(ctx context.Context) ([]storage.Event, error) {
	args := m.Called(ctx)
	var events []storage.Event
	if args.Get(0) != nil {
		events = args.Get(0).([]storage.Event)
	}
	return events, args.Error(1)
}

func (m *Storage) GetEvent(ctx context.Context, uuid uuid.UUID) (storage.Event, error) {
	args := m.Called(ctx, uuid)
	return args.Get(0).(storage.Event), args.Error(1)
}

func (m *Storage) GetEventsByTimeRange(ctx context.Context, startTime, endTime time.Time) ([]storage.Event, error) {
	args := m.Called(ctx, startTime, endTime)
	return args.Get(0).([]storage.Event), args.Error(1)
}
