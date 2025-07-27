package internalhttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/app"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/logger"
)

type Server struct {
	server *http.Server
	logger *logger.Logger
	app    *app.App
}

type Logger interface{}

func NewServer(address string, app *app.App) *Server {
	handler := NewHandler(app)
	return &Server{
		app:    app,
		logger: app.Logger,
		server: &http.Server{
			Addr:         address,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      LoggingMiddleware(handler.mux, handler.logger),
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	go func() {
		s.logger.Info("HTTP server listening", "address", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error(err.Error())
		}
	}()

	<-ctx.Done()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("HTTP server Shutdown.")

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("HTTP server Shutdown error:" + err.Error())
	}

	return nil
}
