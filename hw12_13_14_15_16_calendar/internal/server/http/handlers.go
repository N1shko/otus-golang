package internalhttp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/app"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/logger"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
)

type Handler struct {
	logger *logger.Logger
	app    *app.App
	mux    *http.ServeMux
}

func (h *Handler) Router() http.Handler {
	return LoggingMiddleware(h.mux, h.logger) // Use LoggingMiddleware from middleware.go
}

type Application interface {
	CreateEvent(context.Context, storage.Event) error
}

func NewHandler(logger *logger.Logger, app *app.App) *Handler {
	mux := http.NewServeMux()
	handler := &Handler{logger, app, mux}
	mux.HandleFunc("/hello", HelloHandler)
	return handler
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	fmt.Fprintf(w, "Hello, World!")
}
