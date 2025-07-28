package internalhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/app"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/logger"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
	mockstorage "github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage/mock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type handlerFunc func(*Handler, http.ResponseWriter, *http.Request)

type Test struct {
	name           string
	body           string
	method         string
	uri            string
	callFunc       handlerFunc
	setupMock      func(*mockstorage.Storage)
	expectedStatus int
	expectedBody   string
	expectedEvents []storage.Event
}

var tests = []Test{
	// AddEvent
	{
		name:     "successful creation",
		method:   http.MethodPost,
		uri:      "/events",
		callFunc: (*Handler).createEvent,
		body: `{
							"title": "Team Meeting",
							"dateStart": "2025-08-01T10:00:00Z",
							"dateEnd": "2025-08-01T11:00:00Z",
							"description": "Weekly team sync to discuss project progress",
							"userId": "123e4567-e89b-12d3-a456-426614174000",
							"sendBefore": "2025-07-31T10:00:00Z"
							}`,
		setupMock: func(m *mockstorage.Storage) {
			m.On("AddEvent", mock.Anything, mock.AnythingOfType("storage.Event")).Return(nil).Once()
		},
		expectedStatus: http.StatusOK,
		expectedBody:   "Successfully created event",
	},
	{
		name:           "invalid JSON body",
		method:         http.MethodPost,
		uri:            "/events",
		callFunc:       (*Handler).createEvent,
		body:           `{invalid json}`,
		setupMock:      func(_ *mockstorage.Storage) {},
		expectedStatus: http.StatusBadRequest,
		expectedBody:   "Invalid request body\n",
	},
	{
		name:           "missing required fields",
		method:         http.MethodPost,
		uri:            "/events",
		callFunc:       (*Handler).createEvent,
		body:           `{"title":""}`,
		setupMock:      func(_ *mockstorage.Storage) {},
		expectedStatus: http.StatusBadRequest,
		expectedBody:   "Missing required fields\n",
	},
	{
		name:     "end date before start date",
		method:   http.MethodPost,
		uri:      "/events",
		callFunc: (*Handler).createEvent,
		body: `{
							"title": "Team Meeting",
							"dateStart": "2025-08-01T10:00:00Z",
							"dateEnd": "2025-08-01T9:00:00Z",
							"description": "Weekly team sync to discuss project progress",
							"userId": "123e4567-e89b-12d3-a456-426614174000",
							"sendBefore": "2025-07-31T10:00:00Z"
							}`,
		setupMock:      func(_ *mockstorage.Storage) {},
		expectedStatus: http.StatusBadRequest,
		expectedBody:   "End date must be after start date\n",
	},
	// ListEvent
	{
		name:     "successful retrieval list event",
		method:   http.MethodGet,
		uri:      "/events",
		callFunc: (*Handler).listEvents,
		setupMock: func(m *mockstorage.Storage) {
			events := []storage.Event{
				{
					ID:          uuid.New(),
					Title:       "Event 1",
					DateStart:   time.Now(),
					DateEnd:     time.Now().Add(time.Hour),
					Description: "Something",
					UserID:      "123e4567-e89b-12d3-a456-426614174000",
					SendBefore:  time.Now().Add(time.Minute * 3),
				},
				{
					ID:          uuid.New(),
					Title:       "Event 2",
					DateStart:   time.Now().Add(2 * time.Hour),
					DateEnd:     time.Now().Add(3 * time.Hour),
					Description: "Something",
					UserID:      "123e4567-e89b-12d3-a456-426614173000",
					SendBefore:  time.Now().Add(time.Minute * 3),
				},
			}
			m.On("ListEvents", mock.Anything).Return(events, nil).Once()
		},
		expectedStatus: http.StatusOK,
		expectedEvents: []storage.Event{
			{ID: uuid.New(), Title: "Event 1", DateStart: time.Now(), DateEnd: time.Now().Add(time.Hour)},
			{ID: uuid.New(), Title: "Event 2", DateStart: time.Now().Add(2 * time.Hour), DateEnd: time.Now().Add(3 * time.Hour)},
		},
	},
	// DeleteEvent
	{
		name:     "deleteEvent: successful deletion",
		callFunc: (*Handler).deleteEvent,
		method:   http.MethodDelete,
		uri:      "/events/550e8400-e29b-41d4-a716-446655440000",
		body:     "",
		setupMock: func(m *mockstorage.Storage) {
			m.On("DeleteEvent", mock.Anything, mock.MatchedBy(func(e storage.Event) bool {
				return e.ID == uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
			})).Return(nil).Once()
		},
		expectedStatus: http.StatusOK,
		expectedBody:   "Successfully deleted event 550e8400-e29b-41d4-a716-446655440000",
	},
	// UpdateEvent
	{
		name:     "updateEvent: successful update",
		callFunc: (*Handler).updateEvent,
		method:   http.MethodPut,
		uri:      "/events/550e8400-e29b-41d4-a716-446655440000",
		body: `{
    	"title": "New title"
		}`,
		setupMock: func(m *mockstorage.Storage) {
			currentTime := time.Now()
			existingEvent := storage.Event{
				ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				Title:       "Original Event",
				DateStart:   currentTime,
				DateEnd:     currentTime.Add(time.Hour),
				Description: "Original desc",
				UserID:      "original_user",
				SendBefore:  currentTime.Add(time.Minute * 3),
			}
			m.On("GetEvent", mock.Anything,
				uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")).Return(existingEvent, nil).Once()

			m.On("UpdateEvent", mock.Anything, mock.MatchedBy(func(e storage.Event) bool {
				return e.ID == uuid.MustParse("550e8400-e29b-41d4-a716-446655440000") &&
					e.Title == "New title"
			})).Return(nil).Once()
		},
		expectedStatus: http.StatusOK,
		expectedBody:   "Successfully updated event 550e8400-e29b-41d4-a716-446655440000",
	},
	// listEventsBy
	{
		name:     "listEventsBy: successful day range",
		callFunc: (*Handler).listEventsBy,
		method:   http.MethodGet,
		uri:      "/events?range=day&date=2025-07-28",
		body:     "",
		setupMock: func(m *mockstorage.Storage) {
			startTime, _ := time.Parse("2006-01-02", "2025-07-28")
			endTime := startTime.Add(24 * time.Hour)
			events := []storage.Event{
				{ID: uuid.New(), Title: "Event 1", DateStart: startTime, DateEnd: startTime.Add(time.Hour)},
			}
			m.On("GetEventsByTimeRange", mock.Anything, startTime, endTime).Return(events, nil).Once()
		},
		expectedStatus: http.StatusOK,
		expectedEvents: []storage.Event{
			{ID: uuid.New(), Title: "Event 1"},
		},
	},
}

func TestHandler(t *testing.T) {
	logger := logger.New("DEBUG")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mockstorage.New()
			tt.setupMock(mockStorage)
			appInstance := app.New(logger, mockStorage)
			handler := &Handler{
				app:    appInstance,
				logger: logger,
			}
			req := httptest.NewRequest(tt.method, tt.uri, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			tt.callFunc(handler, w, req)

			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			bodyBytes, _ := io.ReadAll(resp.Body)
			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, string(bodyBytes), tt.expectedBody)
				if tt.expectedEvents != nil {
					var actualEvents []storage.Event
					err := json.Unmarshal(bodyBytes, &actualEvents)
					assert.NoError(t, err)
					assert.Equal(t, len(tt.expectedEvents), len(actualEvents))
				}
			} else {
				assert.Equal(t, tt.expectedBody, string(bodyBytes))
			}
			mockStorage.AssertExpectations(t)
		})
	}
}
