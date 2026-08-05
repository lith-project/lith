package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	// Locate testdata relative to this file
	_, thisFile, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(thisFile), "testdata")

	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStderr string // substring match; empty means expect nothing
	}{
		{
			name:       "no args",
			args:       nil,
			wantCode:   2,
			wantStderr: "Usage:",
		},
		{
			name:       "missing --config flag",
			args:       []string{},
			wantCode:   2,
			wantStderr: "Usage:",
		},
		{
			name:       "missing file",
			args:       []string{"--config", "/nonexistent.yaml"},
			wantCode:   2,
			wantStderr: "file not found",
		},
		{
			name:       "malformed yaml",
			args:       []string{"--config", filepath.Join(testdata, "bad.yaml")},
			wantCode:   2,
			wantStderr: "line ",
		},
		{
			name:       "valid config with JSON logging",
			args:       []string{"--config", filepath.Join(testdata, "valid-log.yaml")},
			wantCode:   0,
			wantStderr: "", // stderr has JSON log lines; validated below
		},
		{
			name:       "invalid log format",
			args:       []string{"--config", filepath.Join(testdata, "bad-log.yaml")},
			wantCode:   2,
			wantStderr: "log.format",
		},
		{
			name:       "noop watcher disabled",
			args:       []string{"--config", filepath.Join(testdata, "noop-watcher.yaml")},
			wantCode:   0,
			wantStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			got := run(tt.args, &stdout, &stderr)

			if got != tt.wantCode {
				t.Errorf("run() = %d, want %d", got, tt.wantCode)
			}

			if tt.wantStderr == "" {
				if tt.wantCode == 0 {
					// For success cases, stdout should be empty
					if stdout.Len() > 0 {
						t.Errorf("stdout = %q, want empty", stdout.String())
					}
				}
			} else {
				if !bytes.Contains(stderr.Bytes(), []byte(tt.wantStderr)) {
					t.Errorf("stderr = %q, want to contain %q", stderr.String(), tt.wantStderr)
				}
			}
		})
	}
}

func TestRunJSONLogRecords(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(thisFile), "testdata")

	var stdout, stderr bytes.Buffer
	got := run([]string{"--config", filepath.Join(testdata, "valid-log.yaml")}, &stdout, &stderr)

	if got != 0 {
		t.Fatalf("run() = %d, want 0", got)
	}

	if stdout.Len() > 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}

	// Parse JSON log lines from stderr
	lines := bytes.Split(stderr.Bytes(), []byte("\n"))
	var records []map[string]interface{}
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var rec map[string]interface{}
		if err := json.Unmarshal(line, &rec); err != nil {
			t.Fatalf("failed to parse JSON log line: %v\nline: %s", err, line)
		}
		records = append(records, rec)
	}

	if len(records) != 4 {
		t.Fatalf("got %d log records, want 4", len(records))
	}

	// Record 1: daemon.starting
	if msg, ok := records[0]["msg"]; !ok || msg != "daemon.starting" {
		t.Errorf("record 1 msg = %v, want \"daemon.starting\"", records[0]["msg"])
	}

	// Record 2: config.loaded with vault_path
	if msg, ok := records[1]["msg"]; !ok || msg != "config.loaded" {
		t.Errorf("record 2 msg = %v, want \"config.loaded\"", records[1]["msg"])
	}
	vp, ok := records[1]["vault_path"]
	if !ok {
		t.Errorf("record 2 missing vault_path field")
	} else {
		vpStr, ok := vp.(string)
		if !ok || vpStr == "" {
			t.Errorf("record 2 vault_path = %v, want non-empty string", vp)
		} else if !strings.HasSuffix(vpStr, "test-vault") {
			t.Errorf("record 2 vault_path = %v, want path ending with \"test-vault\"", vp)
		} else if !filepath.IsAbs(vpStr) {
			t.Errorf("record 2 vault_path = %v, want absolute path", vp)
		}
	}

	// Record 3: vault.watching
	if msg, ok := records[2]["msg"]; !ok || msg != "vault.watching" {
		t.Errorf("record 3 msg = %v, want \"vault.watching\"", records[2]["msg"])
	}

	// Record 4: debounce.bounds with quiet and max_delay
	if msg, ok := records[3]["msg"]; !ok || msg != "debounce.bounds" {
		t.Errorf("record 4 msg = %v, want \"debounce.bounds\"", records[3]["msg"])
	}
	if quiet, ok := records[3]["quiet"]; !ok || quiet != "200ms" {
		t.Errorf("record 4 quiet = %v, want \"200ms\"", records[3]["quiet"])
	}
	if maxDelay, ok := records[3]["max_delay"]; !ok || maxDelay != "5s" {
		t.Errorf("record 4 max_delay = %v, want \"5s\"", records[3]["max_delay"])
	}
}

// TestRunDebounceOverride verifies that overriding debounce bounds in config
// changes the effective bounds logged at startup. This proves the composition
// root reads config and passes bounds explicitly.
func TestRunDebounceOverride(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(thisFile), "testdata")

	var stdout, stderr bytes.Buffer
	got := run([]string{"--config", filepath.Join(testdata, "override-debounce.yaml")}, &stdout, &stderr)

	if got != 0 {
		t.Fatalf("run() = %d, want 0", got)
	}

	// Parse JSON log lines from stderr
	lines := bytes.Split(stderr.Bytes(), []byte("\n"))
	var records []map[string]interface{}
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var rec map[string]interface{}
		if err := json.Unmarshal(line, &rec); err != nil {
			t.Fatalf("failed to parse JSON log line: %v\nline: %s", err, line)
		}
		records = append(records, rec)
	}

	// Find the debounce.bounds record
	var boundsRec map[string]interface{}
	for _, rec := range records {
		if msg, ok := rec["msg"]; ok && msg == "debounce.bounds" {
			boundsRec = rec
			break
		}
	}
	if boundsRec == nil {
		t.Fatalf("no debounce.bounds log record found in %d records", len(records))
	}

	// Verify overridden bounds: quiet=500ms, max_delay=10s
	if quiet, ok := boundsRec["quiet"]; !ok || quiet != "500ms" {
		t.Errorf("debounce.bounds quiet = %v, want \"500ms\"", boundsRec["quiet"])
	}
	if maxDelay, ok := boundsRec["max_delay"]; !ok || maxDelay != "10s" {
		t.Errorf("debounce.bounds max_delay = %v, want \"10s\"", boundsRec["max_delay"])
	}
}
