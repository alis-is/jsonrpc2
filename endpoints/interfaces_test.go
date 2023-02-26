package endpoints

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultLogger(t *testing.T) {
	var logger ILogger = &DefaultLogger{}
	assert.NotPanics(t, func() {
		logger.Tracef("Trace")
		logger.Debugf("Debug")
		logger.Infof("Info")
		logger.Warnf("Warn")
		logger.Errorf("Error")
		logger.Fatalf("Fatal")
	})
}
