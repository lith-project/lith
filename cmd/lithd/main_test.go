package main

import (
	"bytes"
	"path/filepath"
	"runtime"
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
			name:       "valid config",
			args:       []string{"--config", filepath.Join(testdata, "valid.yaml")},
			wantCode:   0,
			wantStderr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			got := run(tt.args, &stderr)

			if got != tt.wantCode {
				t.Errorf("run() = %d, want %d", got, tt.wantCode)
			}

			if tt.wantStderr == "" {
				if stderr.Len() > 0 {
					t.Errorf("stderr = %q, want empty", stderr.String())
				}
			} else {
				if !bytes.Contains(stderr.Bytes(), []byte(tt.wantStderr)) {
					t.Errorf("stderr = %q, want to contain %q", stderr.String(), tt.wantStderr)
				}
			}
		})
	}
}
