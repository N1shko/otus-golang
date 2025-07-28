package storage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrExists   = errors.New("storage: this event already exists")
	ErrNotFound = errors.New("storage: could not find that event")
)

type Event struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Title       string    `db:"title" json:"title"`
	DateStart   time.Time `db:"date_start" json:"dateStart"`
	DateEnd     time.Time `db:"date_end" json:"dateEnd"`
	Description string    `db:"descr" json:"description"`
	UserID      string    `db:"user_id" json:"userId"`
	SendBefore  time.Time `db:"send_before" json:"sendBefore"`
}

type EventRepo interface {
	AddEvent(context.Context, Event) error
	UpdateEvent(context.Context, Event) error
	DeleteEvent(context.Context, Event) error
	ListEvents(context.Context) ([]Event, error)
	GetEvent(context.Context, uuid.UUID) (Event, error)
	GetEventsByTimeRange(context.Context, time.Time, time.Time) ([]Event, error)
}
