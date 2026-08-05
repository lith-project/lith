package logging

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/lith-project/lith/internal/core/config"
)

func TestEventUniqueness(t *testing.T) {
	events := []string{
		EventDaemonStarting, EventDaemonStarted, EventConfigLoaded,
		EventVaultWatching, EventFileChanged, EventWatcherGap,
		EventShutdownBegin, EventShutdownDone, EventError,
	}
	seen := make(map[string]string) // value → first constant name
	for _, v := range events {
		if prev, ok := seen[v]; ok {
			t.Errorf("duplicate event value %q: first seen in %s, now again", v, prev)
		}
		seen[v] = v
	}

	attrs := []string{
		AttrVaultPath, AttrPath, AttrOp, AttrCode,
		AttrCause, AttrSignal, AttrDuration, AttrCount,
	}
	seen = make(map[string]string)
	for _, v := range attrs {
		if prev, ok := seen[v]; ok {
			t.Errorf("duplicate attribute value %q: first seen in %s, now again", v, prev)
		}
		seen[v] = v
	}
}

func TestEventFieldNames(t *testing.T) {
	events := map[string]string{
		EventDaemonStarting: AttrVaultPath,
		EventDaemonStarted:  AttrVaultPath,
		EventConfigLoaded:   AttrVaultPath,
		EventVaultWatching:  AttrVaultPath,
		EventFileChanged:    AttrPath,
		EventWatcherGap:     AttrPath,
		EventShutdownBegin:  AttrCause,
		EventShutdownDone:   AttrDuration,
		EventError:          AttrCode,
	}

	for event, attrKey := range events {
		t.Run(event, func(t *testing.T) {
			var buf bytes.Buffer
			logger, err := New(&buf, config.Log{Level: "debug", Format: "json"})
			if err != nil {
				t.Fatalf("New() error: %v", err)
			}
			logger.Info(event, attrKey, "test-value")

			var parsed map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
				t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
			}
			if parsed["msg"] != event {
				t.Errorf("msg = %q, want %q", parsed["msg"], event)
			}
			if parsed[attrKey] != "test-value" {
				t.Errorf("%s = %q, want %q", attrKey, parsed[attrKey], "test-value")
			}
		})
	}
}
