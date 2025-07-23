package sqlstorage

import (
	"context"
	"database/sql"
	"time"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
)

type Storage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) AddEvent(ctx context.Context, e storage.Event) error {
	sql := `INSERT INTO events (id, title, date_start, date_end, descr, user_id, send_before)
            VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := s.db.ExecContext(ctx, sql,
		e.ID, e.Title, e.DateStart, e.DateEnd, e.Description, e.UserID, e.SendBefore)
	return err
}

func (s *Storage) UpdateEvent(ctx context.Context, e storage.Event) error {
	query := `UPDATE events
		SET title = $2, date_start = $3, date_end = $4, descr = $5, user_id = $6, send_before = $7
		WHERE id = $1`
	id := e.ID.String()
	res, err := s.db.ExecContext(ctx, query,
		id, e.Title, e.DateStart, e.DateEnd, e.Description, e.UserID, e.SendBefore,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, e storage.Event) error {
	query := `DELETE FROM events WHERE id = $1`
	res, err := s.db.ExecContext(ctx, query, e.ID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func (s *Storage) ListEvents(ctx context.Context) ([]storage.Event, error) {
	query := `SELECT id, title, date_start, date_end, descr, user_id, send_before
		FROM events`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []storage.Event
	for rows.Next() {
		var e storage.Event
		err := rows.Scan(
			&e.ID, &e.Title, &e.DateStart, &e.DateEnd, &e.Description, &e.UserID, &e.SendBefore,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (s *Storage) GetEvent(ctx context.Context, id uuid.UUID) (storage.Event, error) {
	query := `SELECT id, title, date_start, date_end, descr, user_id, send_before
		FROM events
		WHERE id = $1`
	var e storage.Event
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID, &e.Title, &e.DateStart, &e.DateEnd, &e.Description, &e.UserID, &e.SendBefore,
	)
	if err != nil {
		return storage.Event{}, err
	}
	return e, nil
}

func (s *Storage) GetEventsByTimeRange(
	ctx context.Context,
	dateStart time.Time,
	dateEnd time.Time,
) ([]storage.Event, error) {
	query := `SELECT id, title, date_start, date_end, descr, user_id, send_before
		FROM events
		WHERE date_start >= $1 AND date_start <= $2`
	rows, err := s.db.QueryContext(ctx, query, dateStart, dateEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []storage.Event
	for rows.Next() {
		var e storage.Event
		err := rows.Scan(
			&e.ID, &e.Title, &e.DateStart, &e.DateEnd, &e.Description, &e.UserID, &e.SendBefore,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
