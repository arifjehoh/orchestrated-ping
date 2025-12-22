package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/arifjehoh/orchestrated-ping/internal/models"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	logger    *slog.Logger
	startTime time.Time
}

func New(logger *slog.Logger, startTime time.Time) *Handler {
	return &Handler{
		logger:    logger,
		startTime: startTime,
	}
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("ping request received",
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	response := models.Response{
		Status:  "success",
		Message: "pong",
		Time:    time.Now(),
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(h.startTime).String()

	h.logger.Debug("health check",
		slog.String("uptime", uptime),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	response := models.HealthResponse{
		Status: "healthy",
		Uptime: uptime,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("readiness check",
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	response := models.Response{
		Status:  "ready",
		Message: "application is ready to serve traffic",
		Time:    time.Now(),
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *Handler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response",
			slog.String("error", err.Error()),
		)
	}
}
