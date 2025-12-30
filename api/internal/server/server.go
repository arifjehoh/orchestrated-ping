package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/arifjehoh/orchestrated-ping/internal/config"
	"github.com/arifjehoh/orchestrated-ping/internal/handlers"
	"github.com/arifjehoh/orchestrated-ping/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func New(cfg *config.Config, logger *slog.Logger, handler *handlers.Handler) *Server {
	router := setupRouter(logger, handler)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return &Server{
		httpServer: srv,
		logger:     logger,
	}
}

func setupRouter(logger *slog.Logger, handler *handlers.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Metrics())
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	r.Get("/ping", handler.Ping)
	r.Get("/health", handler.Health)
	r.Get("/ready", handler.Ready)
	r.Handle("/metrics", promhttp.Handler())

	return r
}

func (s *Server) Start() error {
	s.logger.Info("starting server",
		slog.String("address", s.httpServer.Addr),
	)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down server")
	return s.httpServer.Shutdown(ctx)
}
