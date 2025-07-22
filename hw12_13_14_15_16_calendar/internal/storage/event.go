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
	DateStart   time.Time `db:"date_start" json:"date_start"`
	DateEnd     time.Time `db:"date_end" json:"date_end"`
	Description string    `db:"descr" json:"description"`
	UserID      string    `db:"user_id" json:"user_id"`
	SendBefore  time.Time `db:"send_before" json:"send_before"`
}

type EventRepo interface {
	AddEvent(context.Context, Event) error
	UpdateEvent(context.Context, Event) error
	DeleteEvent(context.Context, Event) error
	ListEvents(context.Context) ([]Event, error)
	GetEvent(context.Context, uuid.UUID) (Event, error)
}
