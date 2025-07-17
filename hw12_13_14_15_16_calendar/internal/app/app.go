package app

import (
	"context"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/logger"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
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
}

func New(logger *logger.Logger, storage Storage) *App {
	return &App{
		logger,
		storage,
	}
}

func (a *App) CreateEvent(_ context.Context, _, _ string) error {
	// TODO
	return nil
	// return a.storage.CreateEvent(storage.Event{ID: id, Title: title})
}

// TODO
