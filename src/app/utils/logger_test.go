package utils

import (
	"testing"
)

func TestGetLogger(t *testing.T) {
	// Test that GetLogger returns a non-nil logger
	logger1 := GetLogger()
	if logger1 == nil {
		t.Errorf("GetLogger() returned nil, expected non-nil logger")
	}

	// Test that GetLogger returns the same logger instance on multiple calls
	logger2 := GetLogger()
	if logger1 != logger2 {
		t.Errorf("GetLogger() returned different instances, expected the same instance")
	}
}
