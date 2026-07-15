package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestNew_JSONOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Service: "byteport-test",
		Version: "1.0.0",
		Env:     "production",
		Level:   LevelInfo,
		Format:  FormatJSON,
		Output:  &buf,
	})
	logger.Info("hello", "key", "value")

	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("expected JSON output, got: %s (err: %v)", buf.String(), err)
	}
	if out["msg"] != "hello" {
		t.Errorf("expected msg=hello, got %v", out["msg"])
	}
	if out["service"] != "byteport-test" {
		t.Errorf("expected service=byteport-test, got %v", out["service"])
	}
	if out["version"] != "1.0.0" {
		t.Errorf("expected version=1.0.0, got %v", out["version"])
	}
	if out["env"] != "production" {
		t.Errorf("expected env=production, got %v", out["env"])
	}
	if out["level"] != "INFO" {
		t.Errorf("expected level=INFO, got %v", out["level"])
	}
	if out["key"] != "value" {
		t.Errorf("expected key=value, got %v", out["key"])
	}
}

func TestNew_TextOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Service: "byteport-test",
		Version: "1.0.0",
		Env:     "development",
		Level:   LevelDebug,
		Format:  FormatText,
		Output:  &buf,
	})
	logger.Debug("debug-msg", "k", "v")

	out := buf.String()
	if !strings.Contains(out, "msg=debug-msg") {
		t.Errorf("expected text format with msg=debug-msg, got: %s", out)
	}
	if !strings.Contains(out, "k=v") {
		t.Errorf("expected text format with k=v, got: %s", out)
	}
}

func TestNew_Defaults(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{Output: &buf})
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
	logger.Info("ping")
	if buf.Len() == 0 {
		t.Error("expected log output, got empty buffer")
	}
}

func TestNew_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{
		Level:  LevelWarn,
		Format: FormatJSON,
		Output: &buf,
	})
	logger.Info("should-not-appear")
	logger.Warn("should-appear")

	out := buf.String()
	if strings.Contains(out, "should-not-appear") {
		t.Errorf("info should be filtered at warn level, got: %s", out)
	}
	if !strings.Contains(out, "should-appear") {
		t.Errorf("warn should be emitted, got: %s", out)
	}
}

func TestParseLevel(t *testing.T) {
	cases := map[string]Level{
		"debug": LevelDebug,
		"DEBUG": LevelDebug,
		"info":  LevelInfo,
		"warn":  LevelWarn,
		"warning": LevelWarn,
		"error": LevelError,
		"err":   LevelError,
		"":      LevelInfo,
		"unknown": LevelInfo,
	}
	for in, want := range cases {
		if got := ParseLevel(in); got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParseFormat(t *testing.T) {
	cases := map[string]Format{
		"json":    FormatJSON,
		"text":    FormatText,
		"console": FormatText,
		"":        FormatJSON,
		"unknown": FormatJSON,
	}
	for in, want := range cases {
		if got := ParseFormat(in); got != want {
			t.Errorf("ParseFormat(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestWithRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{Format: FormatJSON, Output: &buf})
	logger.With("request_id", "abc-123").Info("hi")

	out := buf.String()
	if !strings.Contains(out, "abc-123") {
		t.Errorf("expected request_id in log, got: %s", out)
	}
}

func TestFromContext_Default(t *testing.T) {
	got := FromContext(context.Background())
	if got == nil {
		t.Error("expected non-nil default logger")
	}
}

func TestWithLogger_RoundTrip(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{Format: FormatJSON, Output: &buf})
	ctx := WithLogger(context.Background(), logger)

	got := FromContext(ctx)
	got.Info("via-context")

	if !strings.Contains(buf.String(), "via-context") {
		t.Errorf("expected 'via-context' in output, got: %s", buf.String())
	}
}

func TestWithTrace(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{Format: FormatJSON, Output: &buf})
	logger.With("trace_id", "0123456789abcdef0123456789abcdef", "span_id", "0123456789abcdef").Info("hello")
	out := buf.String()
	if !strings.Contains(out, "trace_id") || !strings.Contains(out, "span_id") {
		t.Errorf("expected trace_id and span_id, got: %s", out)
	}
}

func TestWithUser(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{Format: FormatJSON, Output: &buf})
	logger.With("user_id", "u-123", "org_id", "o-456").Info("auth")
	out := buf.String()
	if !strings.Contains(out, "u-123") || !strings.Contains(out, "o-456") {
		t.Errorf("expected user_id and org_id, got: %s", out)
	}
}

func TestDefaultHandlerType(t *testing.T) {
	var buf bytes.Buffer
	// Force JSON via Env=production
	logger := New(Config{Env: "production", Output: &buf})
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
	logger.Info("x")
	// JSON output starts with "{"
	if !strings.HasPrefix(strings.TrimSpace(buf.String()), "{") {
		t.Errorf("expected JSON output for production env, got: %s", buf.String())
	}
}

func TestConcurrentLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Config{Format: FormatJSON, Output: &buf})

	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func(n int) {
			logger.Info("concurrent", "n", n)
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 50; i++ {
		<-done
	}
	if buf.Len() == 0 {
		t.Error("expected concurrent output, got empty")
	}
}

func TestLoggerImplementsInterface(t *testing.T) {
	var buf bytes.Buffer
	var _ *slog.Logger = New(Config{Output: &buf})
}