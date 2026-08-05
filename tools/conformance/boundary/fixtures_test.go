package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestFixtures runs the boundary checker binary against each test fixture
// and asserts the expected exit code and violation assertion ID.
func TestFixtures(t *testing.T) {
	// Build the checker binary once into a temp directory.
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "boundary")

	build := exec.Command("go", "build", "-o", binPath, ".")
	// CWD is already the boundary package dir (tools/conformance/boundary/) when go test runs.
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("failed to build checker binary: %v", err)
	}

	// Read the source denylist (relative to this test file's directory).
	denylistSrc, err := filepath.Abs(filepath.Join("..", "core-dependency-denylist.txt"))
	if err != nil {
		t.Fatalf("resolving denylist source path: %v", err)
	}
	denylistContent, err := os.ReadFile(denylistSrc)
	if err != nil {
		t.Fatalf("reading denylist source: %v", err)
	}

	type testCase struct {
		name        string
		assertionID string // expected in output for violations; empty for clean
		exitCode    int
	}

	cases := []testCase{
		{"c1-unclassified", "C-1", 1},
		{"c2-core-imports-adapter", "C-2", 1},
		{"c4-denied-dependency", "C-4", 1},
		{"c5-imports-cmd", "C-5", 1},
		{"clean", "", 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fixtureDir := filepath.Join("testdata", tc.name)

			// Create denylist at <fixture>/tools/conformance/core-dependency-denylist.txt
			// so the checker (run with CWD=fixtureDir) can resolve it via filepath.Abs.
			denylistDir := filepath.Join(fixtureDir, "tools", "conformance")
			if err := os.MkdirAll(denylistDir, 0o755); err != nil {
				t.Fatalf("creating denylist dir: %v", err)
			}
			denylistDst := filepath.Join(denylistDir, "core-dependency-denylist.txt")
			if err := os.WriteFile(denylistDst, denylistContent, 0o644); err != nil {
				t.Fatalf("writing denylist copy: %v", err)
			}
			t.Cleanup(func() {
				os.RemoveAll(filepath.Join(fixtureDir, "tools"))
			})

			// Run the checker binary with CWD = fixture directory.
			cmd := exec.Command(binPath)
			cmd.Dir = fixtureDir
			out, err := cmd.CombinedOutput()

			actualExit := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					actualExit = exitErr.ExitCode()
				} else {
					t.Fatalf("unexpected error running checker: %v", err)
				}
			}

			if actualExit != tc.exitCode {
				t.Errorf("exit code = %d, want %d\noutput:\n%s", actualExit, tc.exitCode, out)
			}

			if tc.exitCode != 0 && tc.assertionID != "" {
				if !strings.Contains(string(out), tc.assertionID) {
					t.Errorf("output missing assertion %q:\n%s", tc.assertionID, out)
				}
			}
		})
	}
}
