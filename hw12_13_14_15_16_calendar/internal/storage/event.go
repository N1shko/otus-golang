package storage

import (
	"context"
	"errors"
	"time"
)

var (
	ErrExists   = errors.New("storage: this event already exists")
	ErrNotFound = errors.New("storage: could not find that event")
)

type Event struct {
	ID          string    `db:"id"`
	Title       string    `db:"title"`
	DateStart   time.Time `db:"date_start"`
	DateEnd     time.Time `db:"date_end"`
	Description string    `db:"descr"`
	UserID      string    `db:"user_id"`
	SendBefore  time.Time `db:"send_before"`
}

type EventRepo interface {
	AddEvent(context.Context, Event) error
	UpdateEvent(context.Context, Event) error
	DeleteEvent(context.Context, Event) error
	ListEvents(context.Context) ([]Event, error)
}
