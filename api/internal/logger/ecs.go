package logger

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/arifjehoh/orchestrated-ping/internal/config"
)

type ECSHandler struct {
	handler     slog.Handler
	serviceName string
	version     string
}

func NewECSHandler(w io.Writer, serviceName, version string) *ECSHandler {
	return &ECSHandler{
		handler: slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
		serviceName: serviceName,
		version:     version,
	}
}

func (h *ECSHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *ECSHandler) Handle(ctx context.Context, r slog.Record) error {
	attrs := make(map[string]interface{})

	attrs["@timestamp"] = r.Time.UTC().Format(time.RFC3339Nano)
	attrs["ecs.version"] = config.ECSVersion
	attrs["message"] = r.Message
	attrs["log.level"] = r.Level.String()

	attrs["service.name"] = h.serviceName
	attrs["service.version"] = h.version

	r.Attrs(func(a slog.Attr) bool {
		h.mapAttribute(attrs, a.Key, a.Value.Any())
		return true
	})

	b, err := json.Marshal(attrs)
	if err != nil {
		return err
	}

	os.Stdout.Write(b)
	os.Stdout.Write([]byte("\n"))

	return nil
}

func (h *ECSHandler) mapAttribute(attrs map[string]interface{}, key string, val interface{}) {
	switch key {
	case "method":
		attrs["http.request.method"] = val
	case "path":
		attrs["url.path"] = val
	case "status":
		attrs["http.response.status_code"] = val
	case "bytes":
		attrs["http.response.body.bytes"] = val
	case "duration":
		if d, ok := val.(time.Duration); ok {
			attrs["event.duration"] = d.Nanoseconds()
		} else {
			attrs["event.duration"] = val
		}
	case "remote_addr":
		attrs["client.address"] = val
	case "request_id":
		attrs["trace.id"] = val
	case "error":
		attrs["error.message"] = val
	case "uptime":
		attrs["event.uptime"] = val
	case "port":
		attrs["server.port"] = val
	case "environment":
		attrs["service.environment"] = val
	default:
		attrs[key] = val
	}
}

func (h *ECSHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ECSHandler{
		handler:     h.handler.WithAttrs(attrs),
		serviceName: h.serviceName,
		version:     h.version,
	}
}

func (h *ECSHandler) WithGroup(name string) slog.Handler {
	return &ECSHandler{
		handler:     h.handler.WithGroup(name),
		serviceName: h.serviceName,
		version:     h.version,
	}
}

func New(cfg *config.Config) *slog.Logger {
	handler := NewECSHandler(os.Stdout, cfg.Service.Name, cfg.Service.Version)
	return slog.New(handler)
}
