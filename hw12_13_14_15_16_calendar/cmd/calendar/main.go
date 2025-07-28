package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/app"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/logger"
	internalgrpc "github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/server/grpc"
	internalhttp "github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/server/http"
	"github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/N1shko/otus-golang/hw12_13_14_15_calendar/internal/storage/sql"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := NewConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to parse config from %s", configFile)
	}
	logg := logger.New(config.Logger.Level)
	var repo storage.EventRepo
	ctx := context.Background()
	switch config.Storage.Type {
	case "memory":
		repo = memorystorage.New()

	case "db":
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Storage.DB.Host,
			config.Storage.DB.Port,
			config.Storage.DB.User,
			config.Storage.DB.Password,
			config.Storage.DB.Database,
		)

		db, err := sql.Open("pgx", dsn)
		if err != nil {
			log.Fatalf("failed to open DB: %v", err)
		}

		ctx, DBcancel := context.WithTimeout(ctx, 5*time.Second)
		if err := db.PingContext(ctx); err != nil {
			DBcancel()
			log.Fatalf("cannot ping DB: %v", err)
		}
		defer DBcancel()

		logg.Info("Successfully connected to database")
		repo = sqlstorage.NewPostgresStorage(db)

	default:
		log.Fatalf("unsupported storage type: %s", config.Storage.Type) //nolint:gocritic
	}
	calendar := app.New(logg, repo)
	httpServer := internalhttp.NewServer(":"+config.Server.Port.HTTP, calendar)
	grpcServer := internalgrpc.NewServer(":"+config.Server.Port.GRPC, calendar)
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	// Starting both servers
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := httpServer.Start(ctx); err != nil {
			logg.Error("HTTP server failed", "error", err.Error())
		}
	}()

	go func() {
		defer wg.Done()
		if err := grpcServer.Start(ctx); err != nil {
			logg.Error("gRPC server failed", "error", err.Error())
		}
	}()
	logg.Info("calendar is running...")
	<-ctx.Done() // Waiting for cancellation
	logg.Info("stopping the calendar...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	var shutdownWG sync.WaitGroup
	shutdownWG.Add(2)
	// Gracefully shutting down both servers
	go func() {
		defer shutdownWG.Done()
		if err := httpServer.Stop(shutdownCtx); err != nil {
			logg.Error("HTTP server shutdown error", "error", err.Error())
		}
	}()

	go func() {
		defer shutdownWG.Done()
		if err := grpcServer.Stop(shutdownCtx); err != nil {
			logg.Error("gRPC server shutdown error", "error", err.Error())
		}
	}()

	shutdownWG.Wait()
	logg.Info("Application shutdown complete")
}
