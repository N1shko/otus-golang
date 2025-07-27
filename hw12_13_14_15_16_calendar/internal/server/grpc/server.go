package internalgrpc

import (
	"context"
	"errors"
	"net"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/app"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/logger"
	pb "github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/server/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	pb.UnimplementedEventServiceServer
	server  *grpc.Server
	logger  *logger.Logger
	app     *app.App
	address string
}

func NewServer(addr string, app *app.App) *Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor((&Server{app: app}).LoggingUnaryInterceptor),
	)
	s := &Server{
		server:  grpcServer,
		address: addr,
		app:     app,
		logger:  app.Logger,
	}

	pb.RegisterEventServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	return s
}

func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		s.logger.Error("failed to listen", "error", err.Error())
		return err
	}

	go func() {
		s.logger.Info("gRPC server listening", "address", s.address)
		if err := s.server.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			s.logger.Error("gRPC server failed to serve", "error", err.Error())
		}
	}()

	<-ctx.Done()

	s.logger.Info("gRPC server shutdown initiated")
	s.server.GracefulStop()
	s.logger.Info("gRPC server stopped")

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("gRPC server shutdown requested")

	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		s.logger.Warn("gRPC server forceful stop due to context cancellation")
		s.server.Stop()
		return ctx.Err()
	}

	return nil
}
