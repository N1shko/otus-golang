package internalhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/app"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/logger"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
	"github.com/google/uuid"
)

type Handler struct {
	logger *logger.Logger
	app    *app.App
	mux    *http.ServeMux
}

type EventPatch struct {
	Title       string    `json:"title,omitempty"`
	DateStart   time.Time `json:"dateStart,omitempty"`
	DateEnd     time.Time `json:"dateEnd,omitempty"`
	Description string    `json:"description,omitempty"`
	UserID      string    `json:"userId,omitempty"`
	SendBefore  time.Time `json:"sendBefore,omitempty"`
}

func (h *Handler) Router() http.Handler {
	return LoggingMiddleware(h.mux, h.logger)
}

type Application interface {
	CreateEvent(context.Context, storage.Event) error
}

func NewHandler(logger *logger.Logger, app *app.App) *Handler {
	mux := http.NewServeMux()
	handler := &Handler{logger, app, mux}
	mux.HandleFunc("/hello", HelloHandler)
	mux.HandleFunc("/events", handler.handleEvents)
	mux.HandleFunc("/events/", handler.handleEventByID)
	return handler
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, "Hello, World!")
}

func (h *Handler) handleEvents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createEvent(w, r)
	case http.MethodGet:
		h.listEvents(w, r)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleEventByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/events/"):]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPatch, http.MethodPut:
		h.updateEvent(w, r, id)
	case http.MethodDelete:
		h.deleteEvent(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) createEvent(w http.ResponseWriter, r *http.Request) {
	var event storage.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if event.Title == "" || event.DateStart.IsZero() || event.DateEnd.IsZero() {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if event.DateEnd.Before(event.DateStart) {
		http.Error(w, "End date must be after start date", http.StatusBadRequest)
		return
	}

	event.ID = uuid.New()

	if err := h.app.AddEvent(r.Context(), event); err != nil {
		h.logger.Error("Failed to create event", "error", err.Error())
		http.Error(w, "Failed to create event", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully created event %v", event.ID)
}

func (h *Handler) listEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.app.ListEvents(r.Context())
	if err != nil {
		h.logger.Error("Failed to list events", "error", err.Error())
		http.Error(w, "Failed to list events", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) updateEvent(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	existingEvent, err := h.app.GetEvent(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to fetch event", "error", err.Error())
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	var patch EventPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	updatedEvent := existingEvent
	if patch.Title != "" {
		updatedEvent.Title = patch.Title
	}
	if !patch.DateStart.IsZero() {
		updatedEvent.DateStart = patch.DateStart
	}
	if !patch.DateEnd.IsZero() {
		updatedEvent.DateEnd = patch.DateEnd
	}
	if patch.Description != "" {
		updatedEvent.Description = patch.Description
	}
	if patch.UserID != "" {
		updatedEvent.UserID = patch.UserID
	}
	if !patch.SendBefore.IsZero() {
		updatedEvent.SendBefore = patch.SendBefore
	}
	updatedEvent.ID = id
	if updatedEvent.Title == "" || updatedEvent.DateStart.IsZero() || updatedEvent.DateEnd.IsZero() {
		http.Error(w, "Missing required fields after update", http.StatusBadRequest)
		return
	}

	if updatedEvent.DateEnd.Before(updatedEvent.DateStart) {
		http.Error(w, "End date must be after start date", http.StatusBadRequest)
		return
	}

	if err := h.app.UpdateEvent(r.Context(), updatedEvent); err != nil {
		h.logger.Error("Failed to update event", "error", err.Error())
		http.Error(w, "Failed to update event", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully updated event %v", updatedEvent.ID)
}

func (h *Handler) deleteEvent(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	event := storage.Event{ID: id}

	if err := h.app.DeleteEvent(r.Context(), event); err != nil {
		h.logger.Error("Failed to delete event", "error", err.Error())
		http.Error(w, "Failed to delete event", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully deleted event %v", event.ID)
}
