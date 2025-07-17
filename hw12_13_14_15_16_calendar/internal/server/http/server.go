package internalhttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/app"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/logger"
)

type Server struct {
	logger *logger.Logger
	server *http.Server
}

type Logger interface {
}

func NewServer(logger *logger.Logger, address string, app *app.App) *Server {
	handler := NewHandler(logger, app)
	return &Server{
		logger,
		&http.Server{
			Addr:         address,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      handler.Router(),
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Print(err.Error())
		}
	}()

	<-ctx.Done()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	fmt.Print("Server Shutdown.")

	if err := s.server.Shutdown(ctx); err != nil {
		fmt.Print("Server Shutdown error:" + err.Error())
	}

	return nil
}
