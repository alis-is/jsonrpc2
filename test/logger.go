package test

import (
	"context"
	"fmt"
	"log/slog"
)

type testLoggerHandler struct {
	collected  string
	lastLog    string
	logChannel chan string
}

// NewTestLogger creates a new TestLogger.
func newTestLoggerHandler() *testLoggerHandler {
	return &testLoggerHandler{
		logChannel: make(chan string, 100), // Buffer size can be adjusted
	}
}

// Handle collects log entries and sends them to a channel.
func (t *testLoggerHandler) Handle(ctx context.Context, record slog.Record) error {

	context := make(map[string]interface{})

	record.Attrs(func(attr slog.Attr) bool {
		context[attr.Key] = attr.Value.Any()
		return true
	})

	// Format the log entry
	log := fmt.Sprintf("%s: %s", record.Level, record.Message)
	if len(context) > 0 {
		log += fmt.Sprintf(" %v", context)
	}

	// Collect the log
	t.collected += log + "\n"
	t.lastLog = log
	t.logChannel <- log

	return nil
}

func (t *testLoggerHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newLogger := *t
	for _, attr := range attrs {
		newLogger.collected += fmt.Sprintf("%s: %v ", attr.Key, attr.Value)
	}
	return &newLogger
}

// WithGroup returns a new handler with a group context.
func (t *testLoggerHandler) WithGroup(name string) slog.Handler {
	newLogger := *t
	newLogger.collected += fmt.Sprintf("group: %s ", name)
	return &newLogger
}

func (t *testLoggerHandler) Enabled(context.Context, slog.Level) bool {
	return true // Enable all log levels
}

type TestLogger struct {
	*slog.Logger
	handler *testLoggerHandler
}

func NewLogger() *TestLogger {
	handler := newTestLoggerHandler()
	return &TestLogger{
		Logger:  slog.New(handler),
		handler: handler,
	}
}

func (d *TestLogger) Collected() string {
	return d.handler.collected
}

func (d *TestLogger) Prune() {
	d.handler.collected = ""
	d.handler.lastLog = ""
}

func (d *TestLogger) LastLog() string {
	return d.handler.lastLog
}

func (d *TestLogger) LogChannel() <-chan string {
	return d.handler.logChannel
}
