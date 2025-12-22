package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arifjehoh/orchestrated-ping/internal/config"
	"github.com/arifjehoh/orchestrated-ping/internal/handlers"
	"github.com/arifjehoh/orchestrated-ping/internal/logger"
	"github.com/arifjehoh/orchestrated-ping/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg)
	slog.SetDefault(log)

	// Record start time for uptime tracking
	startTime := time.Now()

	// Initialize handlers with dependencies
	handler := handlers.New(log, startTime)

	// Create and start server
	srv := server.New(cfg, log, handler)

	// Log application startup
	log.Info("application starting",
		slog.String("service", cfg.Service.Name),
		slog.String("version", cfg.Service.Version),
		slog.String("environment", cfg.Environment),
		slog.String("port", cfg.Server.Port),
	)

	// Start server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Error("server failed to start", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("received shutdown signal")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("server stopped gracefully")
}
