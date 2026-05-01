package utils

import (
	"os"
	"testing"
)

func TestLoggingSettings(t *testing.T) {
	logFile := "test_webapp.log"
	LoggingSettings(logFile)
	defer os.Remove(logFile)

	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("Log file %s was not created", logFile)
	}
}
