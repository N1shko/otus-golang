package sqlstorage

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
	"github.com/stretchr/testify/require"
)

func makeTestEvent(id string) storage.Event {
	now := time.Now()
	return storage.Event{
		ID:          id,
		Title:       "Test Event",
		DateStart:   now,
		DateEnd:     now.Add(time.Hour),
		Description: "Description",
		UserID:      "user1",
		SendBefore:  now.Add(time.Hour - 10*time.Minute),
	}
}

func TestStorage_AddEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error initializing sqlmock: %s", err)
	}
	defer db.Close()

	s := NewPostgresStorage(db)

	event := storage.Event{
		ID:          "1",
		Title:       "Test Event",
		DateStart:   time.Now(),
		DateEnd:     time.Now().Add(time.Hour),
		Description: "Test description",
		UserID:      "user1",
		SendBefore:  time.Now().Add(time.Hour - time.Minute*10),
	}

	mock.ExpectExec(regexp.QuoteMeta(
		`INSERT INTO events (id, title, date_start, date_end, descr, user_id, send_before)
            VALUES ($1, $2, $3, $4, $5, $6, $7)`)).
		WithArgs(event.ID, event.Title, event.DateStart, event.DateEnd, event.Description, event.UserID, event.SendBefore).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.AddEvent(context.Background(), event)
	require.NoError(t, err)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestStorage_UpdateEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := NewPostgresStorage(db)
	event := makeTestEvent("1")
	event.Title = "Updated Title"

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE events
		SET title = $2, date_start = $3, date_end = $4, descr = $5, user_id = $6, send_before = $7
		WHERE id = $1`)).
		WithArgs(event.ID, event.Title, event.DateStart, event.DateEnd, event.Description, event.UserID, event.SendBefore).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = s.UpdateEvent(context.Background(), event)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestStorage_UpdateEvent_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := NewPostgresStorage(db)
	event := makeTestEvent("1")

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE events
		SET title = $2, date_start = $3, date_end = $4, descr = $5, user_id = $6, send_before = $7
		WHERE id = $1`)).
		WithArgs(event.ID, event.Title, event.DateStart, event.DateEnd, event.Description, event.UserID, event.SendBefore).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = s.UpdateEvent(context.Background(), event)
	require.ErrorIs(t, err, storage.ErrNotFound)
}

func TestStorage_DeleteEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	s := NewPostgresStorage(db)
	event := makeTestEvent("1")

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM events WHERE id = $1`)).
		WithArgs(event.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = s.DeleteEvent(context.Background(), event)
	require.NoError(t, err)
}
