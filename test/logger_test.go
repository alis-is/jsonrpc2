package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestLogger(t *testing.T) {
	logger := NewLogger()
	assert.NotPanics(t, func() {
		logger.Debug("Debug")
		logger.Info("Info")
		logger.Warn("Warn")
		logger.Error("Error")
	})
}

func TestTestLoggerCollected(t *testing.T) {
	logger := NewLogger()
	logger.Debug("Debug")
	logger.Info("Info")
	logger.Warn("Warn")
	logger.Error("Error")
	assert.Contains(t, logger.Collected(), "Debug")
	assert.Contains(t, logger.Collected(), "Info")
	assert.Contains(t, logger.Collected(), "Warn")
	assert.Contains(t, logger.Collected(), "Error")
}

func TestTestLoggerLastLog(t *testing.T) {
	logger := NewLogger()
	logger.Debug("Debug")
	logger.Info("Info")
	logger.Warn("Warn")
	logger.Error("Error")
	assert.Contains(t, logger.LastLog(), "Error")
}

func TestTestLoggerPrune(t *testing.T) {
	logger := NewLogger()
	logger.Debug("Debug")
	logger.Info("Info")
	logger.Warn("Warn")
	logger.Error("Error")
	logger.Prune()
	assert.Equal(t, "", logger.Collected())
}

func TestTestLoggerLastLogChannel(t *testing.T) {
	logger := NewLogger()
	logger.Debug("Debug")
	logger.Info("Info")
	logger.Warn("Warn")
	logger.Error("Error")
	assert.Contains(t, <-logger.LogChannel(), "Debug")
	assert.Contains(t, <-logger.LogChannel(), "Info")
	assert.Contains(t, <-logger.LogChannel(), "Warn")
	assert.Contains(t, <-logger.LogChannel(), "Error")
}
