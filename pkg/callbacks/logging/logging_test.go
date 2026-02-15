package logging

import (
	"bytes"
	"context"
	"log/slog"
	"sync"
	"testing"
)

func TestLoggingDefaults(t *testing.T) {
	l, err := New(Config{})
	if err != nil {
		t.Fatalf("Unable to create new logger - %s", err)
	}

	t.Run("Info Logging", func(t *testing.T) {
		b, err := l.Info([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Error Logging", func(t *testing.T) {
		b, err := l.Error([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Warn Logging", func(t *testing.T) {
		b, err := l.Warn([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Debug Logging", func(t *testing.T) {
		b, err := l.Debug([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Trace Logging", func(t *testing.T) {
		b, err := l.Trace([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})
}

func TestLoggingFunc(t *testing.T) {
	log := NoOpLog{}
	l, err := New(Config{Log: log})
	if err != nil {
		t.Fatalf("Unable to create new logger - %s", err)
	}

	t.Run("Info Logging", func(t *testing.T) {
		b, err := l.Info([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Error Logging", func(t *testing.T) {
		b, err := l.Error([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Warn Logging", func(t *testing.T) {
		b, err := l.Warn([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Debug Logging", func(t *testing.T) {
		b, err := l.Debug([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})

	t.Run("Trace Logging", func(t *testing.T) {
		b, err := l.Trace([]byte("Testing"))
		if err != nil || len(b) > 0 {
			t.Errorf("Invalid return from logger function - %s", err)
		}
	})
}

// spyLogger is a test implementation that captures log calls
type spyLogger struct {
mu      sync.Mutex
calls   []logCall
}

type logCall struct {
level   string
message string
}

func (s *spyLogger) Info(args ...interface{}) {
s.mu.Lock()
defer s.mu.Unlock()
s.calls = append(s.calls, logCall{level: "info", message: formatArgs(args...)})
}

func (s *spyLogger) Warn(args ...interface{}) {
s.mu.Lock()
defer s.mu.Unlock()
s.calls = append(s.calls, logCall{level: "warn", message: formatArgs(args...)})
}

func (s *spyLogger) Error(args ...interface{}) {
s.mu.Lock()
defer s.mu.Unlock()
s.calls = append(s.calls, logCall{level: "error", message: formatArgs(args...)})
}

func (s *spyLogger) Debug(args ...interface{}) {
s.mu.Lock()
defer s.mu.Unlock()
s.calls = append(s.calls, logCall{level: "debug", message: formatArgs(args...)})
}

func (s *spyLogger) Trace(args ...interface{}) {
s.mu.Lock()
defer s.mu.Unlock()
s.calls = append(s.calls, logCall{level: "trace", message: formatArgs(args...)})
}

func (s *spyLogger) getCalls() []logCall {
s.mu.Lock()
defer s.mu.Unlock()
return append([]logCall{}, s.calls...)
}

func (s *spyLogger) reset() {
s.mu.Lock()
defer s.mu.Unlock()
s.calls = nil
}

func formatArgs(args ...interface{}) string {
if len(args) == 0 {
return ""
}
if len(args) == 1 {
return args[0].(string)
}
result := ""
for _, arg := range args {
result += arg.(string)
}
return result
}

// TestLoggerBehavior validates that Logger adapter methods call the correct logger methods
func TestLoggerBehavior(t *testing.T) {
spy := &spyLogger{}
logger, err := New(Config{Log: spy})
if err != nil {
t.Fatalf("Failed to create logger: %v", err)
}

tests := []struct {
name          string
logFunc       func([]byte) ([]byte, error)
input         string
expectedLevel string
}{
{
name:          "Info calls Info method",
logFunc:       logger.Info,
input:         "info message",
expectedLevel: "info",
},
{
name:          "Warn calls Warn method",
logFunc:       logger.Warn,
input:         "warn message",
expectedLevel: "warn",
},
{
name:          "Error calls Error method",
logFunc:       logger.Error,
input:         "error message",
expectedLevel: "error",
},
{
name:          "Debug calls Debug method",
logFunc:       logger.Debug,
input:         "debug message",
expectedLevel: "debug",
},
{
name:          "Trace calls Trace method",
logFunc:       logger.Trace,
input:         "trace message",
expectedLevel: "trace",
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
spy.reset()

_, err := tt.logFunc([]byte(tt.input))
if err != nil {
t.Errorf("Unexpected error: %v", err)
}

calls := spy.getCalls()
if len(calls) != 1 {
t.Fatalf("Expected 1 call, got %d", len(calls))
}

if calls[0].level != tt.expectedLevel {
t.Errorf("Expected level %q, got %q", tt.expectedLevel, calls[0].level)
}

if calls[0].message != tt.input {
t.Errorf("Expected message %q, got %q", tt.input, calls[0].message)
}
})
}
}

// spySlogHandler is a test handler that captures slog calls
type spySlogHandler struct {
mu      sync.Mutex
records []slogRecord
}

type slogRecord struct {
level   slog.Level
message string
}

func (h *spySlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
return true
}

func (h *spySlogHandler) Handle(ctx context.Context, r slog.Record) error {
h.mu.Lock()
defer h.mu.Unlock()
h.records = append(h.records, slogRecord{
level:   r.Level,
message: r.Message,
})
return nil
}

func (h *spySlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
return h
}

func (h *spySlogHandler) WithGroup(name string) slog.Handler {
return h
}

func (h *spySlogHandler) getRecords() []slogRecord {
h.mu.Lock()
defer h.mu.Unlock()
return append([]slogRecord{}, h.records...)
}

func (h *spySlogHandler) reset() {
h.mu.Lock()
defer h.mu.Unlock()
h.records = nil
}

// TestSlogAdapterBehavior validates that SlogAdapter methods call slog with correct levels
func TestSlogAdapterBehavior(t *testing.T) {
handler := &spySlogHandler{}
slogger := slog.New(handler)
adapter := NewSlogAdapter(slogger)

tests := []struct {
name          string
logFunc       func(...interface{})
input         string
expectedLevel slog.Level
}{
{
name:          "Info logs at Info level",
logFunc:       adapter.Info,
input:         "info message",
expectedLevel: slog.LevelInfo,
},
{
name:          "Warn logs at Warn level",
logFunc:       adapter.Warn,
input:         "warn message",
expectedLevel: slog.LevelWarn,
},
{
name:          "Error logs at Error level",
logFunc:       adapter.Error,
input:         "error message",
expectedLevel: slog.LevelError,
},
{
name:          "Debug logs at Debug level",
logFunc:       adapter.Debug,
input:         "debug message",
expectedLevel: slog.LevelDebug,
},
{
name:          "Trace logs at Trace level",
logFunc:       adapter.Trace,
input:         "trace message",
expectedLevel: LevelTrace,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
handler.reset()

tt.logFunc(tt.input)

records := handler.getRecords()
if len(records) != 1 {
t.Fatalf("Expected 1 record, got %d", len(records))
}

if records[0].level != tt.expectedLevel {
t.Errorf("Expected level %v, got %v", tt.expectedLevel, records[0].level)
}

if records[0].message != tt.input {
t.Errorf("Expected message %q, got %q", tt.input, records[0].message)
}
})
}
}

// TestSlogAdapterWithMultipleArgs validates that SlogAdapter formats multiple args correctly
func TestSlogAdapterWithMultipleArgs(t *testing.T) {
handler := &spySlogHandler{}
slogger := slog.New(handler)
adapter := NewSlogAdapter(slogger)

adapter.Info("part1", " ", "part2")

records := handler.getRecords()
if len(records) != 1 {
t.Fatalf("Expected 1 record, got %d", len(records))
}

expected := "part1 part2"
if records[0].message != expected {
t.Errorf("Expected message %q, got %q", expected, records[0].message)
}
}

// TestSlogAdapterTraceLevelValue validates that trace level is below debug
func TestSlogAdapterTraceLevelValue(t *testing.T) {
if LevelTrace >= slog.LevelDebug {
t.Errorf("Expected LevelTrace (%d) to be less than LevelDebug (%d)", LevelTrace, slog.LevelDebug)
}
}

// TestSlogAdapterIntegration validates end-to-end behavior with actual slog
func TestSlogAdapterIntegration(t *testing.T) {
var buf bytes.Buffer
handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
Level: LevelTrace,
})
slogger := slog.New(handler)
adapter := NewSlogAdapter(slogger)

// Test each level
adapter.Info("info test")
adapter.Warn("warn test")
adapter.Error("error test")
adapter.Debug("debug test")
adapter.Trace("trace test")

output := buf.String()

// Verify each message appears in output
expectedMessages := []string{"info test", "warn test", "error test", "debug test", "trace test"}
for _, msg := range expectedMessages {
if !bytes.Contains([]byte(output), []byte(msg)) {
t.Errorf("Expected output to contain %q, got: %s", msg, output)
}
}
}
