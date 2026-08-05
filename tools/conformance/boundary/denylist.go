package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// denylistEntry represents a single line from the denylist file.
type denylistEntry struct {
	prefix string
}

// loadDenylist reads and parses the denylist file.
// Returns error if file is unreadable (this is a fatal error, not a silent pass).
func loadDenylist(path string) ([]denylistEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening denylist: %w", err)
	}
	// Close error is intentionally ignored: the scanner reports read errors
	// and the file is opened read-only.
	defer func() { _ = f.Close() }()

	var entries []denylistEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries = append(entries, denylistEntry{prefix: line})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading denylist: %w", err)
	}
	return entries, nil
}

// coreDepGraph holds the parsed output of go list -deps -json ./internal/core/...
type coreDepGraph struct {
	packages map[string]depGraphPkg
}

type depGraphPkg struct {
	importPath string
	modulePath string
	imports    []string
}

// goListDepPkg mirrors the JSON structure emitted by "go list -deps -json".
type goListDepPkg struct {
	ImportPath string   `json:"ImportPath"`
	Imports    []string `json:"Imports"`
	Module     *struct {
		Path string `json:"Path"`
	} `json:"Module"`
}

// buildCoreDepGraph runs "go list -deps -json ./internal/core/..." and
// parses the JSON stream into a graph.
// If no Go packages exist under internal/core/ (exit code from go list),
// returns an empty graph (not an error).
func buildCoreDepGraph(modulePrefix string) (*coreDepGraph, error) {
	graph := &coreDepGraph{packages: make(map[string]depGraphPkg)}

	cmd := exec.Command("go", "list", "-deps", "-json", modulePrefix+"/internal/core/...")
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("creating stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting go list: %w", err)
	}

	dec := json.NewDecoder(stdout)
	for dec.More() {
		var pkg goListDepPkg
		if err := dec.Decode(&pkg); err != nil {
			return nil, fmt.Errorf("decoding go list output: %w", err)
		}

		var modulePath string
		if pkg.Module != nil {
			modulePath = pkg.Module.Path
		}

		graph.packages[pkg.ImportPath] = depGraphPkg{
			importPath: pkg.ImportPath,
			modulePath: modulePath,
			imports:    pkg.Imports,
		}
	}

	if err := cmd.Wait(); err != nil {
		// "no Go files" or "no packages" is not an error — just no core to check.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() != 0 {
			return graph, nil
		}
		return nil, fmt.Errorf("go list failed: %w", err)
	}

	return graph, nil
}

// checkDenylist takes the graph and denylist, finds all C-4 violations.
func checkDenylist(graph *coreDepGraph, denylist []denylistEntry) []Violation {
	var violations []Violation

	for importPath := range graph.packages {
		if !isCorePackage(importPath) {
			continue
		}

		for _, dep := range denylist {
			chain := traceChain(graph, importPath, dep.prefix)
			if chain != nil {
				violations = append(violations, Violation{
					Assertion: "C-4",
					Package:   importPath,
					Detail:    fmt.Sprintf("core package transitively imports denied module (prefix %s); chain: %s", dep.prefix, strings.Join(chain, " -> ")),
				})
			}
		}
	}

	return violations
}

// isCorePackage returns true if the import path is under internal/core/.
func isCorePackage(importPath string) bool {
	idx := strings.Index(importPath, "/internal/core")
	if idx < 0 {
		return false
	}
	rest := importPath[idx+len("/internal/core"):]
	return rest == "" || strings.HasPrefix(rest, "/")
}

// traceChain does BFS from start through the graph's imports,
// returning the path (slice of import paths) to target, or nil if unreachable.
func traceChain(graph *coreDepGraph, start, targetPrefix string) []string {
	type node struct {
		importPath string
		path       []string
	}

	visited := make(map[string]bool)
	queue := []node{{importPath: start, path: []string{start}}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.importPath] {
			continue
		}
		visited[current.importPath] = true

		pkg, ok := graph.packages[current.importPath]
		if !ok {
			continue
		}

		if pkg.modulePath != "" && strings.HasPrefix(pkg.modulePath, targetPrefix) {
			return current.path
		}

		for _, imp := range pkg.imports {
			if !visited[imp] {
				newPath := make([]string, len(current.path), len(current.path)+1)
				copy(newPath, current.path)
				newPath = append(newPath, imp)
				queue = append(queue, node{importPath: imp, path: newPath})
			}
		}
	}

	return nil
}
