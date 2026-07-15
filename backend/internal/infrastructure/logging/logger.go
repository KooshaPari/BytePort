// Package logging provides structured logging for BytePort.
//
// It wraps log/slog (Go 1.21+ stdlib) with project conventions:
//   - JSON handler in production, text handler in development.
//   - Consistent base attributes (service, version, env).
//   - Convenience helpers for common fields (request_id, trace_id, span_id).
//
// Usage:
//
//	logger := logging.New(logging.Config{
//	    Service: "byteport-api",
//	    Version: version.Version,
//	    Level:   logging.LevelInfo,
//	    Format:  logging.FormatJSON,
//	})
//	logger = logger.WithRequestID(req.Header.Get("X-Request-ID"))
//	logger.Info("deployment started", "deployment_id", id)
package logging

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Level aliases mirror slog.Level for ergonomics.
type Level = slog.Level

const (
	LevelDebug Level = slog.LevelDebug
	LevelInfo  Level = slog.LevelInfo
	LevelWarn  Level = slog.LevelWarn
	LevelError Level = slog.LevelError
)

// Format selects the slog handler type.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Config configures a new structured Logger.
type Config struct {
	// Service name injected as the "service" attribute on every record.
	Service string
	// Version injected as the "version" attribute on every record.
	Version string
	// Env injected as the "env" attribute (production|staging|development|test).
	Env string
	// Level is the minimum log level emitted (default: LevelInfo).
	Level Level
	// Format selects the handler (json|text). Default: json when Env=="production" else text.
	Format Format
	// Output is the destination for log records (default: os.Stderr).
	Output io.Writer
	// AddSource adds the calling file:line to each record (default: false in prod).
	AddSource bool
}

// New builds a Logger from Config with sensible defaults.
func New(cfg Config) *slog.Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stderr
	}
	if cfg.Env == "" {
		cfg.Env = detectEnv()
	}
	if cfg.Service == "" {
		cfg.Service = "byteport"
	}
	if cfg.Format == "" {
		if cfg.Env == "production" {
			cfg.Format = FormatJSON
		} else {
			cfg.Format = FormatText
		}
	}
	if cfg.Level == 0 {
		cfg.Level = LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource || cfg.Env == "development",
	}

	var handler slog.Handler
	switch cfg.Format {
	case FormatText:
		handler = slog.NewTextHandler(cfg.Output, opts)
	default:
		handler = slog.NewJSONHandler(cfg.Output, opts)
	}

	logger := slog.New(handler).With(
		"service", cfg.Service,
		"version", cfg.Version,
		"env", cfg.Env,
	)
	return logger
}

// ParseLevel converts a string (debug|info|warn|error) into a Level.
// Unknown strings default to LevelInfo.
func ParseLevel(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug
	case "warn", "warning":
		return LevelWarn
	case "error", "err":
		return LevelError
	default:
		return LevelInfo
	}
}

// ParseFormat converts a string (json|text) into a Format.
// Unknown strings default to FormatJSON.
func ParseFormat(s string) Format {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "text", "console":
		return FormatText
	default:
		return FormatJSON
	}
}

func detectEnv() string {
	if env := os.Getenv("BYTEPORT_ENV"); env != "" {
		return env
	}
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		return env
	}
	if env := os.Getenv("NODE_ENV"); env != "" {
		return env
	}
	if env := os.Getenv("GO_ENV"); env != "" {
		return env
	}
	return "development"
}

// ---- convenience helpers -------------------------------------------------

// FromContext returns the Logger stored in ctx (via WithLogger).
// If absent, returns slog.Default().
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}

// WithLogger returns a new context carrying logger.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// WithRequestID returns a new Logger annotated with request_id.
func WithRequestID(parent *slog.Logger, requestID string) *slog.Logger {
	if requestID == "" {
		return parent
	}
	return parent.With("request_id", requestID)
}

// WithTrace returns a new Logger annotated with OpenTelemetry trace_id and span_id.
func WithTrace(parent *slog.Logger, traceID, spanID string) *slog.Logger {
	if traceID == "" && spanID == "" {
		return parent
	}
	return parent.With("trace_id", traceID, "span_id", spanID)
}

// WithUser returns a new Logger annotated with user_id and org_id.
func WithUser(parent *slog.Logger, userID, orgID string) *slog.Logger {
	args := make([]any, 0, 4)
	if userID != "" {
		args = append(args, "user_id", userID)
	}
	if orgID != "" {
		args = append(args, "org_id", orgID)
	}
	if len(args) == 0 {
		return parent
	}
	return parent.With(args...)
}

type loggerKey struct{}