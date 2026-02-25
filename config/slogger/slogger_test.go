package slogger

import (
	"bytes"
	"context"
	"flag"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestLogLevelToString(t *testing.T) {
	tests := []struct {
		level slog.Level
		want  string
	}{
		{LevelTrace, "TRACE"},
		{slog.Level(-9), "TRACE-1"},
		{slog.Level(-12), "TRACE-4"},
		{slog.LevelDebug, "DEBUG"},
		{slog.LevelInfo, "INFO"},
		{slog.LevelWarn, "WARN"},
		{slog.LevelError, "ERROR"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := LogLevelToString(tt.level)
			if got != tt.want {
				t.Errorf("LogLevelToString(%d) = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}

func TestLevelValueString(t *testing.T) {
	// Nil LevelVar should default to INFO.
	lv := LevelValue{}
	if got := lv.String(); got != "INFO" {
		t.Errorf("nil LevelValue.String() = %q, want %q", got, "INFO")
	}

	// Set to debug.
	lv = LevelValue{LevelVar: &slog.LevelVar{}}
	lv.Set("debug")
	if got := lv.String(); got != "DEBUG" {
		t.Errorf("LevelValue.String() = %q, want %q", got, "DEBUG")
	}
}

func TestLevelValueSet(t *testing.T) {
	lv := LevelValue{LevelVar: &slog.LevelVar{}}

	if err := lv.Set("warn"); err != nil {
		t.Fatalf("Set(warn) error: %v", err)
	}
	if lv.Level() != slog.LevelWarn {
		t.Errorf("level = %v, want %v", lv.Level(), slog.LevelWarn)
	}

	if err := lv.Set("invalid-level"); err == nil {
		t.Error("Set(invalid-level) should return error")
	}
}

func TestSlogConfigRegisterFlags(t *testing.T) {
	cfg := &SlogConfig{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg.RegisterFlags(fs)

	if err := fs.Parse([]string{"-slogger.log-level", "warn"}); err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if cfg.LogLevel != "warn" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "warn")
	}
}

func TestSlogConfigValidate(t *testing.T) {
	cfg := &SlogConfig{LogLevel: "info"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() error for valid level: %v", err)
	}

	cfg = &SlogConfig{LogLevel: "not-a-level"}
	if err := cfg.Validate(); err == nil {
		t.Error("Validate() should return error for invalid level")
	}
}

func TestSlogConfigMerge(t *testing.T) {
	cfg := &SlogConfig{LogLevel: "debug"}
	other := &SlogConfig{LogLevel: "error"}

	if err := cfg.Merge(other); err != nil {
		t.Fatalf("Merge error: %v", err)
	}
	if cfg.LogLevel != "error" {
		t.Errorf("after merge, LogLevel = %q, want %q", cfg.LogLevel, "error")
	}

	// Empty string should not override.
	cfg = &SlogConfig{LogLevel: "warn"}
	other = &SlogConfig{LogLevel: ""}
	if err := cfg.Merge(other); err != nil {
		t.Fatalf("Merge error: %v", err)
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("after merge with empty, LogLevel = %q, want %q", cfg.LogLevel, "warn")
	}
}

func TestGetLoggerDefault(t *testing.T) {
	var buf bytes.Buffer
	cfg := &SlogConfig{LogLevel: "info"}

	logger := cfg.GetLogger(WithOutput(&buf))
	logger.Info("hello world")

	output := buf.String()
	if !strings.Contains(output, "hello world") {
		t.Errorf("expected log output to contain 'hello world', got: %s", output)
	}
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected log output to contain 'INFO', got: %s", output)
	}
}

func TestGetLoggerWithJSONHandler(t *testing.T) {
	var buf bytes.Buffer
	cfg := &SlogConfig{LogLevel: "debug"}

	logger := cfg.GetLogger(WithOutput(&buf), WithJSONHandler())
	logger.Debug("test message")

	output := buf.String()
	// JSON handler outputs JSON, should contain the message in JSON format.
	if !strings.Contains(output, `"msg":"test message"`) {
		t.Errorf("expected JSON output with msg field, got: %s", output)
	}
}

func TestGetLoggerWithTextHandler(t *testing.T) {
	var buf bytes.Buffer
	cfg := &SlogConfig{LogLevel: "debug"}

	logger := cfg.GetLogger(WithOutput(&buf), WithTextHandler())
	logger.Debug("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("expected text output with message, got: %s", output)
	}
	if !strings.Contains(output, "DEBUG") {
		t.Errorf("expected text output with DEBUG level, got: %s", output)
	}
}

func TestGetLoggerWithHandlerFactory(t *testing.T) {
	var buf bytes.Buffer
	cfg := &SlogConfig{LogLevel: "info"}

	// Use a custom handler factory that returns a JSON handler.
	logger := cfg.GetLogger(
		WithOutput(&buf),
		WithHandlerFactory(func(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
			return slog.NewJSONHandler(w, opts)
		}),
	)
	logger.Info("factory test")

	output := buf.String()
	if !strings.Contains(output, `"msg":"factory test"`) {
		t.Errorf("expected JSON from factory, got: %s", output)
	}
}

func TestGetLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	cfg := &SlogConfig{LogLevel: "warn"}

	logger := cfg.GetLogger(WithOutput(&buf))
	logger.Info("should not appear")
	logger.Warn("should appear")

	output := buf.String()
	if strings.Contains(output, "should not appear") {
		t.Error("info message should be filtered at warn level")
	}
	if !strings.Contains(output, "should appear") {
		t.Errorf("warn message should appear, got: %s", output)
	}
}

func TestGetLoggerTraceLevel(t *testing.T) {
	var buf bytes.Buffer
	cfg := &SlogConfig{LogLevel: "DEBUG-4"} // slog parses DEBUG-4 as level -8 = TRACE

	logger := cfg.GetLogger(WithOutput(&buf))
	logger.Log(context.Background(), LevelTrace, "trace message")

	output := buf.String()
	if !strings.Contains(output, "TRACE") {
		t.Errorf("expected TRACE level in output, got: %s", output)
	}
	if !strings.Contains(output, "trace message") {
		t.Errorf("expected trace message in output, got: %s", output)
	}
}

func TestGetLoggerWithAddSourceFalse(t *testing.T) {
	var buf bytes.Buffer
	cfg := &SlogConfig{LogLevel: "info"}

	logger := cfg.GetLogger(WithOutput(&buf), WithAddSource(false))
	logger.Info("no source")

	output := buf.String()
	if strings.Contains(output, "slogger_test.go") {
		t.Errorf("source should not appear when AddSource is false, got: %s", output)
	}
}

func TestGetLoggerWithReplaceAttr(t *testing.T) {
	var buf bytes.Buffer
	cfg := &SlogConfig{LogLevel: "info"}

	// Custom ReplaceAttr that removes the time key.
	logger := cfg.GetLogger(
		WithOutput(&buf),
		WithReplaceAttr(func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		}),
	)
	logger.Info("no time")

	output := buf.String()
	if strings.Contains(output, "time=") {
		t.Errorf("time attribute should be removed, got: %s", output)
	}
}

func TestUserInformationHandler(t *testing.T) {
	var buf bytes.Buffer
	cfg := &SlogConfig{LogLevel: "info"}

	logger := cfg.GetLogger(WithOutput(&buf), WithAddSource(false))

	ctx := context.Background()
	ctx = context.WithValue(ctx, UserIDKey{}, "user-123")
	ctx = context.WithValue(ctx, ServerIDKey{}, "server-456")

	logger.InfoContext(ctx, "with user info")

	output := buf.String()
	if !strings.Contains(output, "user-123") {
		t.Errorf("expected userID in output, got: %s", output)
	}
	if !strings.Contains(output, "server-456") {
		t.Errorf("expected serverID in output, got: %s", output)
	}
}

func TestUserInformationHandlerNoContext(t *testing.T) {
	var buf bytes.Buffer
	cfg := &SlogConfig{LogLevel: "info"}

	logger := cfg.GetLogger(WithOutput(&buf), WithAddSource(false))
	logger.Info("no user info")

	output := buf.String()
	if strings.Contains(output, "userID") {
		t.Errorf("userID should not appear without context value, got: %s", output)
	}
	if strings.Contains(output, "serverID") {
		t.Errorf("serverID should not appear without context value, got: %s", output)
	}
}

func TestGetLoggerInvalidLevel(t *testing.T) {
	var buf bytes.Buffer
	cfg := &SlogConfig{LogLevel: "not-valid"}

	// Should not panic; should fall back to debug.
	logger := cfg.GetLogger(WithOutput(&buf))
	logger.Debug("fallback works")

	output := buf.String()
	if !strings.Contains(output, "fallback works") {
		t.Errorf("expected debug message with fallback level, got: %s", output)
	}
}

func TestDynamicLevelChangeViaMerge(t *testing.T) {
	var buf bytes.Buffer
	cfg := &SlogConfig{LogLevel: "info"}

	logger := cfg.GetLogger(WithOutput(&buf), WithAddSource(false))

	// Info should appear at info level.
	logger.Info("visible")
	if !strings.Contains(buf.String(), "visible") {
		t.Fatal("info message should appear at info level")
	}

	buf.Reset()

	// Debug should NOT appear at info level.
	logger.Debug("hidden")
	if strings.Contains(buf.String(), "hidden") {
		t.Fatal("debug message should not appear at info level")
	}

	// Merge to debug level — the existing logger should pick it up.
	if err := cfg.Merge(&SlogConfig{LogLevel: "debug"}); err != nil {
		t.Fatalf("Merge error: %v", err)
	}

	buf.Reset()
	logger.Debug("now visible")
	if !strings.Contains(buf.String(), "now visible") {
		t.Errorf("debug message should appear after merge to debug, got: %s", buf.String())
	}
}

func TestDynamicLevelChangeViaSetLevel(t *testing.T) {
	var buf bytes.Buffer
	cfg := &SlogConfig{LogLevel: "debug"}

	logger := cfg.GetLogger(WithOutput(&buf), WithAddSource(false))

	// Debug should appear.
	logger.Debug("before")
	if !strings.Contains(buf.String(), "before") {
		t.Fatal("debug message should appear at debug level")
	}

	// Change to error level.
	if err := cfg.SetLevel("error"); err != nil {
		t.Fatalf("SetLevel error: %v", err)
	}

	buf.Reset()
	logger.Debug("should be hidden")
	logger.Info("also hidden")
	logger.Warn("also hidden")
	logger.Error("visible error")

	output := buf.String()
	if strings.Contains(output, "should be hidden") {
		t.Error("debug should be filtered at error level")
	}
	if strings.Contains(output, "also hidden") {
		t.Error("info/warn should be filtered at error level")
	}
	if !strings.Contains(output, "visible error") {
		t.Errorf("error should appear at error level, got: %s", output)
	}
}

func TestSetLevelInvalidReturnsError(t *testing.T) {
	cfg := &SlogConfig{LogLevel: "info"}
	cfg.GetLogger() // initialize the LevelVar

	err := cfg.SetLevel("not-a-level")
	if err == nil {
		t.Error("SetLevel with invalid level should return error")
	}
	// Level should remain unchanged.
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel should remain 'info' after invalid SetLevel, got %q", cfg.LogLevel)
	}
}

func TestLevelVarReturnsNilBeforeGetLogger(t *testing.T) {
	cfg := &SlogConfig{LogLevel: "info"}
	if cfg.LevelVar() != nil {
		t.Error("LevelVar() should return nil before GetLogger is called")
	}
}

func TestLevelVarReturnsSameInstance(t *testing.T) {
	cfg := &SlogConfig{LogLevel: "info"}
	var buf bytes.Buffer

	cfg.GetLogger(WithOutput(&buf))
	lv1 := cfg.LevelVar()
	cfg.GetLogger(WithOutput(&buf))
	lv2 := cfg.LevelVar()

	if lv1 != lv2 {
		t.Error("LevelVar() should return the same *slog.LevelVar across GetLogger calls")
	}
}

func TestLevelVarDirectMutation(t *testing.T) {
	var buf bytes.Buffer
	cfg := &SlogConfig{LogLevel: "info"}

	logger := cfg.GetLogger(WithOutput(&buf), WithAddSource(false))

	// Directly mutate the LevelVar.
	cfg.LevelVar().Set(slog.LevelError)

	logger.Info("should be hidden")
	logger.Error("should appear")

	output := buf.String()
	if strings.Contains(output, "should be hidden") {
		t.Error("info should be filtered after LevelVar set to error")
	}
	if !strings.Contains(output, "should appear") {
		t.Errorf("error should appear, got: %s", output)
	}
}

func TestMergeUpdatesLevelVarAfterGetLogger(t *testing.T) {
	cfg := &SlogConfig{LogLevel: "info"}
	var buf bytes.Buffer
	cfg.GetLogger(WithOutput(&buf))

	if cfg.LevelVar().Level() != slog.LevelInfo {
		t.Errorf("initial level should be INFO, got %v", cfg.LevelVar().Level())
	}

	if err := cfg.Merge(&SlogConfig{LogLevel: "warn"}); err != nil {
		t.Fatalf("Merge error: %v", err)
	}

	if cfg.LevelVar().Level() != slog.LevelWarn {
		t.Errorf("level should be WARN after merge, got %v", cfg.LevelVar().Level())
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("LogLevel string should be 'warn' after merge, got %q", cfg.LogLevel)
	}
}
