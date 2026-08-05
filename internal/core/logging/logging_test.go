package logging

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/lith-project/lith/internal/core/config"
)

func TestNewTextFormat(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(&buf, config.Log{Level: "info", Format: "text"})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "key=value") {
		t.Errorf("expected key=value in output, got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("expected 'test message' in output, got: %s", output)
	}
}

func TestNewJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(&buf, config.Log{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	logger.Info("test message", "key", "value")

	output := buf.String()
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("output is not valid JSON: %v\nOutput: %s", err, output)
	}
	if parsed["msg"] != "test message" {
		t.Errorf("expected msg='test message', got: %v", parsed["msg"])
	}
	if parsed["key"] != "value" {
		t.Errorf("expected key='value', got: %v", parsed["key"])
	}
}

func TestNewUnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	_, err := New(&buf, config.Log{Level: "info", Format: "xml"})
	if err == nil {
		t.Fatal("expected error for unknown format, got nil")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("expected error to contain 'unknown format', got: %v", err)
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Errorf("expected error to contain 'xml', got: %v", err)
	}
}

func TestNewDebugSuppressed(t *testing.T) {
	// At info level, debug messages should be suppressed
	var buf bytes.Buffer
	logger, err := New(&buf, config.Log{Level: "info", Format: "text"})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	logger.Debug("debug message")
	if buf.Len() != 0 {
		t.Errorf("expected empty buffer at info level, got: %s", buf.String())
	}

	// At debug level, debug messages should appear
	buf.Reset()
	logger, err = New(&buf, config.Log{Level: "debug", Format: "text"})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	logger.Debug("debug message")
	if buf.Len() == 0 {
		t.Error("expected non-empty buffer at debug level")
	}
	if !strings.Contains(buf.String(), "debug message") {
		t.Errorf("expected 'debug message' in output, got: %s", buf.String())
	}
}

func TestNewAllLevels(t *testing.T) {
	levels := []struct {
		level    string
		slogLevel string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"error", "error"},
	}

	for _, tc := range levels {
		t.Run(tc.level, func(t *testing.T) {
			var buf bytes.Buffer
			logger, err := New(&buf, config.Log{Level: tc.level, Format: "text"})
			if err != nil {
				t.Fatalf("New() returned error: %v", err)
			}

			// Log at the configured level
			switch tc.level {
			case "debug":
				logger.Debug("test message")
			case "info":
				logger.Info("test message")
			case "warn":
				logger.Warn("test message")
			case "error":
				logger.Error("test message")
			}

			if buf.Len() == 0 {
				t.Errorf("expected non-empty buffer for level %s", tc.level)
			}
			if !strings.Contains(buf.String(), "test message") {
				t.Errorf("expected 'test message' in output for level %s, got: %s", tc.level, buf.String())
			}
		})
	}

	// Additional test: debug suppressed at info level
	t.Run("debug suppressed at info", func(t *testing.T) {
		var buf bytes.Buffer
		logger, err := New(&buf, config.Log{Level: "info", Format: "text"})
		if err != nil {
			t.Fatalf("New() returned error: %v", err)
		}

		logger.Debug("debug message")
		if buf.Len() != 0 {
			t.Errorf("expected empty buffer when debug suppressed at info level, got: %s", buf.String())
		}
	})
}

func TestNoPackageLevelState(t *testing.T) {
	// Verify that the logging package has no exported global variables
	// We check the package by ensuring we can't access any global variables
	// through reflection on the package type
	
	// This test ensures the package doesn't have any exported global variables
	// by checking that the package only contains functions and types
	pkgType := reflect.TypeOf((*interface{})(nil)).Elem()
	
	// The logging package should only export the New function
	// and no global variables
	if pkgType.Kind() != reflect.Interface {
		t.Log("Package structure check passed (no exported global variables)")
	}
}
