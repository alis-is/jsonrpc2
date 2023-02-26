package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestLogger(t *testing.T) {
	logger := NewLogger()
	assert.NotPanics(t, func() {
		logger.Tracef("Trace")
		logger.Debugf("Debug")
		logger.Infof("Info")
		logger.Warnf("Warn")
		logger.Errorf("Error")
		logger.Fatalf("Fatal")
	})
}

func TestTestLoggerCollected(t *testing.T) {
	logger := NewLogger()
	logger.Tracef("Trace")
	logger.Debugf("Debug")
	logger.Infof("Info")
	logger.Warnf("Warn")
	logger.Errorf("Error")
	logger.Fatalf("Fatal")
	assert.Equal(t, "TRACE: Trace\nDEBUG: Debug\nINFO: Info\nWARN: Warn\nERROR: Error\nFATAL: Fatal\n", logger.Collected())
}

func TestTestLoggerLastLog(t *testing.T) {
	logger := NewLogger()
	logger.Tracef("Trace")
	logger.Debugf("Debug")
	logger.Infof("Info")
	logger.Warnf("Warn")
	logger.Errorf("Error")
	logger.Fatalf("Fatal")
	assert.Equal(t, "FATAL: Fatal\n", logger.LastLog())
}

func TestTestLoggerPrune(t *testing.T) {
	logger := NewLogger()
	logger.Tracef("Trace")
	logger.Debugf("Debug")
	logger.Infof("Info")
	logger.Warnf("Warn")
	logger.Errorf("Error")
	logger.Fatalf("Fatal")
	logger.Prune()
	assert.Equal(t, "", logger.Collected())
}

func TestTestLoggerLastLogChannel(t *testing.T) {
	logger := NewLogger()
	logger.Tracef("Trace")
	logger.Debugf("Debug")
	logger.Infof("Info")
	logger.Warnf("Warn")
	logger.Errorf("Error")
	logger.Fatalf("Fatal")
	assert.Contains(t, <-logger.LogChannel(), "TRACE: Trace")
	assert.Contains(t, <-logger.LogChannel(), "DEBUG: Debug")
	assert.Contains(t, <-logger.LogChannel(), "INFO: Info")
	assert.Contains(t, <-logger.LogChannel(), "WARN: Warn")
	assert.Contains(t, <-logger.LogChannel(), "ERROR: Error")
	assert.Contains(t, <-logger.LogChannel(), "FATAL: Fatal")
}
