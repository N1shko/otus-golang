package app

import (
	"context"
	"time"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/logger"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
)

type App struct {
	logger  *logger.Logger
	storage storage.EventRepo
}

type Storage interface {
	AddEvent(context.Context, storage.Event) error
	UpdateEvent(context.Context, storage.Event) error
	DeleteEvent(context.Context, storage.Event) error
	ListEvents(context.Context) ([]storage.Event, error)
	GetEvent(context.Context, uuid.UUID) (storage.Event, error)
	GetEventsByTimeRange(context.Context, time.Time, time.Time) ([]storage.Event, error)
}

func New(logger *logger.Logger, storage Storage) *App {
	return &App{
		logger,
		storage,
	}
}

func (a *App) AddEvent(ctx context.Context, event storage.Event) error {
	return a.storage.AddEvent(ctx, event)
}

func (a *App) UpdateEvent(ctx context.Context, event storage.Event) error {
	return a.storage.UpdateEvent(ctx, event)
}

func (a *App) DeleteEvent(ctx context.Context, event storage.Event) error {
	return a.storage.DeleteEvent(ctx, event)
}

func (a *App) ListEvents(ctx context.Context) ([]storage.Event, error) {
	return a.storage.ListEvents(ctx)
}

func (a *App) GetEvent(ctx context.Context, id uuid.UUID) (storage.Event, error) {
	return a.storage.GetEvent(ctx, id)
}

func (a *App) GetEventsByTimeRange(
	ctx context.Context,
	dateStart time.Time,
	dateEnd time.Time,
) ([]storage.Event, error) {
	return a.storage.GetEventsByTimeRange(ctx, dateStart, dateEnd)
}
