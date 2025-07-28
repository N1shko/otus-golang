package internalgrpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/app"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/logger"
	pb "github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
	mockstorage "github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestServer_ListEvents(t *testing.T) {
	mockStorage := mockstorage.New()
	logger := logger.New("DEBUG")
	appInstance := app.New(logger, mockStorage)

	server := NewServer("localhost:0", appInstance)

	t.Run("successful retrieval", func(t *testing.T) {
		expectedEvents := []storage.Event{
			{ID: uuid.New(), Title: "Event 1", DateStart: time.Now(), DateEnd: time.Now().Add(time.Hour)},
			{ID: uuid.New(), Title: "Event 2", DateStart: time.Now().Add(2 * time.Hour), DateEnd: time.Now().Add(3 * time.Hour)},
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

func TestServer_DeleteEvents(t *testing.T) {
	mockStorage := mockstorage.New()
	logger := logger.New("DEBUG")
	appInstance := app.New(logger, mockStorage)

	server := NewServer("localhost:0", appInstance)

	t.Run("successful delete", func(t *testing.T) {
		eventID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
		req := &pb.DeleteEventRequest{Id: eventID.String()}

		mockStorage.
			On("DeleteEvent", mock.Anything, mock.MatchedBy(func(e storage.Event) bool {
				return e.ID == eventID
			})).
			Return(nil).
			Once()

		ctx := context.Background()
		resp, err := server.DeleteEvent(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		mockStorage.AssertExpectations(t)
	})
}

func TestAddEvent(t *testing.T) {
	now := time.Now()
	start := now.Add(time.Hour)
	end := now.Add(2 * time.Hour)

	tests := []struct {
		name             string
		req              *pb.AddEventRequest
		setupMock        func(m *mockstorage.Storage)
		expectedCode     codes.Code
		expectedErrorMsg string
		validateResponse func(t *testing.T, resp *pb.AddEventResponse)
		cancelContext    bool
	}{
		{
			name: "successful event creation",
			req: &pb.AddEventRequest{
				Title:       "Meeting",
				DateStart:   timestamppb.New(start),
				DateEnd:     timestamppb.New(end),
				Description: "Team sync",
				UserId:      "user-123",
				SendBefore:  timestamppb.New(start.Add(-10 * time.Minute)),
			},
			setupMock: func(m *mockstorage.Storage) {
				m.On("AddEvent", mock.Anything, mock.MatchedBy(func(e storage.Event) bool {
					return e.Title == "Meeting" && e.UserID == "user-123"
				})).Return(nil).Once()
			},
			expectedCode: codes.OK,
		},
		{
			name: "missing required fields",
			req: &pb.AddEventRequest{
				Title: "",
			},
			expectedCode:     codes.InvalidArgument,
			expectedErrorMsg: "Missing required fields",
		},
		{
			name: "end date before start date",
			req: &pb.AddEventRequest{
				Title:     "Event",
				DateStart: timestamppb.New(end),
				DateEnd:   timestamppb.New(start),
			},
			expectedCode:     codes.InvalidArgument,
			expectedErrorMsg: "End date must be after start date",
		},
		{
			name: "app returns error",
			req: &pb.AddEventRequest{
				Title:     "Event",
				DateStart: timestamppb.New(start),
				DateEnd:   timestamppb.New(end),
			},
			setupMock: func(m *mockstorage.Storage) {
				m.On("AddEvent", mock.Anything, mock.Anything).Return(errors.New("db error")).Once()
			},
			expectedCode:     codes.Internal,
			expectedErrorMsg: "Failed to create event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mockstorage.New()
			logger := logger.New("DEBUG")
			appInstance := app.New(logger, mockStorage)
			server := NewServer("localhost:0", appInstance)

			if tt.setupMock != nil {
				tt.setupMock(mockStorage)
			}

			ctx := context.Background()
			if tt.cancelContext {
				c, cancel := context.WithCancel(ctx)
				cancel()
				ctx = c
			}

			resp, err := server.AddEvent(ctx, tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok, "error should be a gRPC status")
				assert.Equal(t, tt.expectedCode, st.Code(), "unexpected gRPC code")
				assert.Equal(t, tt.expectedErrorMsg, st.Message(), "unexpected error message")
				assert.Nil(t, resp)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestUpdateEvent(t *testing.T) {
	now := time.Now()
	start := now.Add(time.Hour)
	end := now.Add(2 * time.Hour)

	existingID := uuid.New()

	existingEvent := storage.Event{
		ID:          existingID,
		Title:       "Old Title",
		DateStart:   start,
		DateEnd:     end,
		Description: "Old Desc",
		UserID:      "user-1",
		SendBefore:  start.Add(-15 * time.Minute),
	}

	tests := []struct {
		name             string
		req              *pb.UpdateEventRequest
		setupMock        func(m *mockstorage.Storage)
		cancelContext    bool
		expectedCode     codes.Code
		expectedErrorMsg string
		validateResponse func(t *testing.T, resp *pb.UpdateEventResponse)
	}{
		{
			name: "successful update",
			req: &pb.UpdateEventRequest{
				Event: &pb.Event{
					Id:          existingID.String(),
					Title:       "Updated Title",
					Description: "Updated Desc",
				},
			},
			setupMock: func(m *mockstorage.Storage) {
				m.On("GetEvent", mock.Anything, existingID).Return(existingEvent, nil).Once()
				m.On("UpdateEvent", mock.Anything, mock.MatchedBy(func(e storage.Event) bool {
					return e.Title == "Updated Title" &&
						e.Description == "Updated Desc" &&
						e.ID == existingID
				})).Return(nil).Once()
			},
			expectedCode: codes.OK,
			validateResponse: func(t *testing.T, resp *pb.UpdateEventResponse) {
				assert.Equal(t, existingID.String(), resp.Id)
			},
		},
		{
			name: "invalid UUID",
			req: &pb.UpdateEventRequest{
				Event: &pb.Event{Id: "not-a-uuid"},
			},
			expectedCode:     codes.InvalidArgument,
			expectedErrorMsg: "Invalid event ID",
		},
		{
			name: "event not found",
			req: &pb.UpdateEventRequest{
				Event: &pb.Event{Id: existingID.String()},
			},
			setupMock: func(m *mockstorage.Storage) {
				m.On("GetEvent", mock.Anything, existingID).Return(storage.Event{}, errors.New("not found")).Once()
			},
			expectedCode:     codes.NotFound,
			expectedErrorMsg: "Event not found",
		},
		{
			name: "missing required fields after update",
			req: &pb.UpdateEventRequest{
				Event: &pb.Event{
					Id:    existingID.String(),
					Title: "",
				},
			},
			setupMock: func(m *mockstorage.Storage) {
				eventWithNoTitle := existingEvent
				eventWithNoTitle.Title = ""
				m.On("GetEvent", mock.Anything, existingID).Return(eventWithNoTitle, nil).Once()
			},
			expectedCode:     codes.InvalidArgument,
			expectedErrorMsg: "Missing required fields after update",
		},
		{
			name: "end date before start date",
			req: &pb.UpdateEventRequest{
				Event: &pb.Event{
					Id:        existingID.String(),
					DateStart: timestamppb.New(end),
					DateEnd:   timestamppb.New(start),
				},
			},
			setupMock: func(m *mockstorage.Storage) {
				m.On("GetEvent", mock.Anything, existingID).Return(existingEvent, nil).Once()
			},
			expectedCode:     codes.InvalidArgument,
			expectedErrorMsg: "End date must be after start date",
		},
		{
			name: "update failure",
			req: &pb.UpdateEventRequest{
				Event: &pb.Event{
					Id:    existingID.String(),
					Title: "Title",
				},
			},
			setupMock: func(m *mockstorage.Storage) {
				m.On("GetEvent", mock.Anything, existingID).Return(existingEvent, nil).Once()
				m.On("UpdateEvent", mock.Anything, mock.Anything).Return(errors.New("db failure")).Once()
			},
			expectedCode:     codes.Internal,
			expectedErrorMsg: "Failed to update event",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mockstorage.New()
			logger := logger.New("DEBUG")
			appInstance := app.New(logger, mockStorage)
			server := NewServer("localhost:0", appInstance)

			if tt.setupMock != nil {
				tt.setupMock(mockStorage)
			}

			ctx := context.Background()
			if tt.cancelContext {
				c, cancel := context.WithCancel(ctx)
				cancel()
				ctx = c
			}

			resp, err := server.UpdateEvent(ctx, tt.req)

			if tt.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if tt.validateResponse != nil {
					tt.validateResponse(t, resp)
				}
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok, "expected gRPC status error")
				assert.Equal(t, tt.expectedCode, st.Code())
				assert.Equal(t, tt.expectedErrorMsg, st.Message())
				assert.Nil(t, resp)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestGetEventByRange(t *testing.T) {
	logger := logger.New("DEBUG")
	type callFunc func(*Server, context.Context, interface{}) (interface{}, error)
	type assertResponse func(t *testing.T, resp interface{}, expectedEvents []storage.Event)
	tests := []struct {
		name              string
		callFunc          callFunc
		request           interface{}
		setupMock         func(*mockstorage.Storage)
		expectedErrorCode codes.Code
		expectedErrorMsg  string
		expectedEvents    []storage.Event
		assertResponse    assertResponse
	}{
		{
			name: "GetEventsByTimeRange: successful day range",
			callFunc: func(s *Server, ctx context.Context, req interface{}) (interface{}, error) {
				return s.GetEventsByTimeRange(ctx, req.(*pb.GetEventsByTimeRangeRequest))
			},
			request: &pb.GetEventsByTimeRangeRequest{
				RangeType: pb.TimeRangeType_DAY,
				StartDate: timestamppb.New(time.Date(2025, 7, 28, 0, 0, 0, 0, time.UTC)),
			},
			setupMock: func(m *mockstorage.Storage) {
				startTime := time.Date(2025, 7, 28, 0, 0, 0, 0, time.UTC)
				endTime := startTime.Add(24 * time.Hour)
				events := []storage.Event{
					{
						ID:        uuid.New(),
						Title:     "Event 1",
						DateStart: startTime,
						DateEnd:   startTime.Add(time.Hour),
					},
				}
				m.On("GetEventsByTimeRange", mock.Anything, startTime, endTime).Return(events, nil).Once()
			},
			expectedErrorCode: codes.OK,
			expectedEvents: []storage.Event{
				{Title: "Event 1"},
			},
			assertResponse: func(t *testing.T, resp interface{}, expectedEvents []storage.Event) {
				getResp, ok := resp.(*pb.GetEventsByTimeRangeResponse)
				assert.True(t, ok, "Response should be GetEventsByTimeRangeResponse")
				assert.Len(t, getResp.Events, len(expectedEvents), "Unexpected number of events")
				for i, pbEvent := range getResp.Events {
					assert.Equal(t, expectedEvents[i].Title, pbEvent.Title, "Event title mismatch")
				}
			},
		},
		{
			name: "GetEventsByTimeRange: successful week range",
			callFunc: func(s *Server, ctx context.Context, req interface{}) (interface{}, error) {
				return s.GetEventsByTimeRange(ctx, req.(*pb.GetEventsByTimeRangeRequest))
			},
			request: &pb.GetEventsByTimeRangeRequest{
				RangeType: pb.TimeRangeType_WEEK,
				StartDate: timestamppb.New(time.Date(2025, 7, 28, 0, 0, 0, 0, time.UTC)),
			},
			setupMock: func(m *mockstorage.Storage) {
				startTime := time.Date(2025, 7, 28, 0, 0, 0, 0, time.UTC)
				endTime := startTime.Add(7 * 24 * time.Hour)
				events := []storage.Event{
					{
						ID:        uuid.New(),
						Title:     "Event 1",
						DateStart: startTime,
						DateEnd:   startTime.Add(time.Hour),
					},
				}
				m.On("GetEventsByTimeRange", mock.Anything, startTime, endTime).Return(events, nil).Once()
			},
			expectedErrorCode: codes.OK,
			expectedEvents: []storage.Event{
				{Title: "Event 1"},
			},
			assertResponse: func(t *testing.T, resp interface{}, expectedEvents []storage.Event) {
				getResp, ok := resp.(*pb.GetEventsByTimeRangeResponse)
				assert.True(t, ok, "Response should be GetEventsByTimeRangeResponse")
				assert.Len(t, getResp.Events, len(expectedEvents), "Unexpected number of events")
				for i, pbEvent := range getResp.Events {
					assert.Equal(t, expectedEvents[i].Title, pbEvent.Title, "Event title mismatch")
				}
				t.Logf("Response events: %v", getResp.Events)
			},
		},
		{
			name: "GetEventsByTimeRange: successful month range",
			callFunc: func(s *Server, ctx context.Context, req interface{}) (interface{}, error) {
				return s.GetEventsByTimeRange(ctx, req.(*pb.GetEventsByTimeRangeRequest))
			},
			request: &pb.GetEventsByTimeRangeRequest{
				RangeType: pb.TimeRangeType_MONTH,
				StartDate: timestamppb.New(time.Date(2025, 7, 28, 0, 0, 0, 0, time.UTC)),
			},
			setupMock: func(m *mockstorage.Storage) {
				startTime := time.Date(2025, 7, 28, 0, 0, 0, 0, time.UTC)
				endTime := startTime.AddDate(0, 1, 0)
				events := []storage.Event{
					{
						ID:        uuid.New(),
						Title:     "Event 1",
						DateStart: startTime,
						DateEnd:   startTime.Add(time.Hour),
					},
				}
				m.On("GetEventsByTimeRange", mock.Anything, startTime, endTime).Return(events, nil).Once()
			},
			expectedErrorCode: codes.OK,
			expectedEvents: []storage.Event{
				{Title: "Event 1"},
			},
			assertResponse: func(t *testing.T, resp interface{}, expectedEvents []storage.Event) {
				getResp, ok := resp.(*pb.GetEventsByTimeRangeResponse)
				assert.True(t, ok, "Response should be GetEventsByTimeRangeResponse")
				assert.Len(t, getResp.Events, len(expectedEvents), "Unexpected number of events")
				for i, pbEvent := range getResp.Events {
					assert.Equal(t, expectedEvents[i].Title, pbEvent.Title, "Event title mismatch")
				}
				t.Logf("Response events: %v", getResp.Events)
			},
		},
		{
			name: "GetEventsByTimeRange: empty start date",
			callFunc: func(s *Server, ctx context.Context, req interface{}) (interface{}, error) {
				return s.GetEventsByTimeRange(ctx, req.(*pb.GetEventsByTimeRangeRequest))
			},
			request: &pb.GetEventsByTimeRangeRequest{
				RangeType: pb.TimeRangeType_DAY,
				StartDate: nil,
			},
			setupMock:         func(m *mockstorage.Storage) {},
			expectedErrorCode: codes.InvalidArgument,
			expectedErrorMsg:  "Date cant be empty",
			assertResponse: func(t *testing.T, resp interface{}, expectedEvents []storage.Event) {
				assert.Nil(t, resp, "Response should be nil on error")
			},
		},
		{
			name: "GetEventsByTimeRange: unspecified range type",
			callFunc: func(s *Server, ctx context.Context, req interface{}) (interface{}, error) {
				return s.GetEventsByTimeRange(ctx, req.(*pb.GetEventsByTimeRangeRequest))
			},
			request: &pb.GetEventsByTimeRangeRequest{
				RangeType: pb.TimeRangeType_TIME_RANGE_TYPE_UNSPECIFIED,
				StartDate: timestamppb.New(time.Date(2025, 7, 28, 0, 0, 0, 0, time.UTC)),
			},
			setupMock:         func(m *mockstorage.Storage) {},
			expectedErrorCode: codes.InvalidArgument,
			expectedErrorMsg:  "Range type must be specified",
			assertResponse: func(t *testing.T, resp interface{}, expectedEvents []storage.Event) {
				assert.Nil(t, resp, "Response should be nil on error")
			},
		},
		{
			name: "GetEventsByTimeRange: invalid date format",
			callFunc: func(s *Server, ctx context.Context, req interface{}) (interface{}, error) {
				return s.GetEventsByTimeRange(ctx, req.(*pb.GetEventsByTimeRangeRequest))
			},
			request: &pb.GetEventsByTimeRangeRequest{
				RangeType: pb.TimeRangeType_DAY,
				StartDate: timestamppb.New(time.Time{}),
			},
			setupMock:         func(m *mockstorage.Storage) {},
			expectedErrorCode: codes.InvalidArgument,
			expectedErrorMsg:  "Invalid date format, use YYYY-MM-DD",
			assertResponse: func(t *testing.T, resp interface{}, expectedEvents []storage.Event) {
				assert.Nil(t, resp, "Response should be nil on error")
			},
		},
		{
			name: "GetEventsByTimeRange: unsupported range type",
			callFunc: func(s *Server, ctx context.Context, req interface{}) (interface{}, error) {
				return s.GetEventsByTimeRange(ctx, req.(*pb.GetEventsByTimeRangeRequest))
			},
			request: &pb.GetEventsByTimeRangeRequest{
				RangeType: 999, // Invalid enum value
				StartDate: timestamppb.New(time.Date(2025, 7, 28, 0, 0, 0, 0, time.UTC)),
			},
			setupMock:         func(m *mockstorage.Storage) {},
			expectedErrorCode: codes.InvalidArgument,
			expectedErrorMsg:  "Unsupported range type",
			assertResponse: func(t *testing.T, resp interface{}, expectedEvents []storage.Event) {
				assert.Nil(t, resp, "Response should be nil on error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mockstorage.New()
			tt.setupMock(mockStorage)
			appInstance := app.New(logger, mockStorage)
			server := NewServer("localhost:0", appInstance)
			ctx := context.Background()
			if tt.expectedErrorCode == codes.Canceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}
			resp, err := tt.callFunc(server, ctx, tt.request)

			if tt.expectedErrorCode == codes.OK {
				assert.NoError(t, err, "Unexpected error")
				assert.NotNil(t, resp, "Response should not be nil")
				tt.assertResponse(t, resp, tt.expectedEvents)
			} else {
				assert.Error(t, err, "Expected an error")
				st, ok := status.FromError(err)
				assert.True(t, ok, "Error should be a gRPC status")
				assert.Equal(t, tt.expectedErrorCode, st.Code(), "Unexpected error code")
				assert.Equal(t, tt.expectedErrorMsg, st.Message(), "Unexpected error message")
				t.Logf("Error: %v", err)
				tt.assertResponse(t, resp, tt.expectedEvents)
			}
			mockStorage.AssertExpectations(t)
		})
	}
}
