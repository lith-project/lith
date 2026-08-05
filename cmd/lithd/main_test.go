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

	if len(records) != 2 {
		t.Fatalf("got %d log records, want 2", len(records))
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
}
