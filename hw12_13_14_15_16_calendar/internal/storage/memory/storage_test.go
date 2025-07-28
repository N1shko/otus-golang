package memorystorage

import (
	"context"
	"testing"
	"time"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func makeTestEvent(id uuid.UUID) storage.Event {
	now := time.Now()
	return storage.Event{
		ID:          id,
		Title:       "Test Event",
		DateStart:   now,
		DateEnd:     now.Add(time.Hour),
		Description: "A description",
		UserID:      "user1",
		SendBefore:  now.Add(time.Hour - 10*time.Minute),
	}
}

func TestAddEvent(t *testing.T) {
	s := New()
	ctx := context.Background()
	e := makeTestEvent(uuid.New())

	err := s.AddEvent(ctx, e)

	require.NoError(t, err)
	err = s.AddEvent(ctx, e)

	require.ErrorIs(t, err, storage.ErrExists)
}

func TestUpdateEvent(t *testing.T) {
	s := New()
	ctx := context.Background()
	unique := uuid.New()
	e := makeTestEvent(unique)
	err := s.UpdateEvent(ctx, e)
	require.ErrorIs(t, err, storage.ErrNotFound)

	_ = s.AddEvent(ctx, e)
	e.Title = "Updated Title"
	err = s.UpdateEvent(ctx, e)
	require.NoError(t, err)

	got, _ := s.GetEvent(ctx, unique)
	require.Equal(t, "Updated Title", got.Title)
}

func TestDeleteEvent(t *testing.T) {
	s := New()
	ctx := context.Background()
	unique := uuid.New()

	e := makeTestEvent(unique)
	err := s.DeleteEvent(ctx, e)
	require.ErrorIs(t, err, storage.ErrNotFound)

	_ = s.AddEvent(ctx, e)
	err = s.DeleteEvent(ctx, e)
	require.NoError(t, err)

	_, err = s.GetEvent(ctx, unique)
	require.ErrorIs(t, err, storage.ErrNotFound)
}

func TestListEvents(t *testing.T) {
	s := New()
	ctx := context.Background()

	events, err := s.ListEvents(ctx)
	require.NoError(t, err)
	require.Empty(t, events)

	_ = s.AddEvent(ctx, makeTestEvent(uuid.New()))
	_ = s.AddEvent(ctx, makeTestEvent(uuid.New()))

	events, err = s.ListEvents(ctx)
	require.NoError(t, err)
	require.Len(t, events, 2)
}
