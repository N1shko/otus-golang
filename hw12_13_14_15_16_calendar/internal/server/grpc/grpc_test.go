package internalgrpc

import (
	"context"
	"errors"
	"testing"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/app"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/logger"
	pb "github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
	mockstorage "github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServer_ListEvents(t *testing.T) {
	mockStorage := mockstorage.New()
	logger := logger.New("DEBUG")
	appInstance := app.New(logger, mockStorage)

	server := NewServer("localhost:0", appInstance)

	t.Run("successful retrieval", func(t *testing.T) {
		expectedEvents := []storage.Event{
			{ID: uuid.New(), Title: "Event 1"},
			{ID: uuid.New(), Title: "Event 2"},
		}
		mockStorage.On("ListEvents", mock.Anything).Return(expectedEvents, nil).Once()

		ctx := context.Background()
		req := &pb.ListEventsRequest{}

		resp, err := server.ListEvents(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Events, len(expectedEvents))
		for i, pbEvent := range resp.Events {
			assert.Equal(t, expectedEvents[i].ID.String(), pbEvent.Id)
			assert.Equal(t, expectedEvents[i].Title, pbEvent.Title)
		}
		mockStorage.AssertExpectations(t)
	})

	// Test error case
	t.Run("error from storage", func(t *testing.T) {
		mockStorage.On("ListEvents", mock.Anything).Return(nil, errors.New("storage error")).Once()

		ctx := context.Background()
		req := &pb.ListEventsRequest{}

		resp, err := server.ListEvents(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "storage error", err.Error())
		mockStorage.AssertExpectations(t)
	})

	t.Run("context cancelled", func(t *testing.T) {
		mockStorage := mockstorage.New()
		appInstance := app.New(logger, mockStorage)
		server := NewServer("localhost:0", appInstance)

		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel()

		req := &pb.ListEventsRequest{}

		resp, err := server.ListEvents(cancelledCtx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, context.Canceled, err)
		mockStorage.AssertNumberOfCalls(t, "ListEvents", 0)
	})
}
