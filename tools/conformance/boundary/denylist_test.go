package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDenylist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "denylist.txt")
	content := `# comment line

github.com/pinecone-io/
github.com/foo/bar
# another comment
github.com/baz/qux

`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	entries, err := loadDenylist(path)
	if err != nil {
		t.Fatalf("loadDenylist returned error: %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].prefix != "github.com/pinecone-io/" {
		t.Errorf("entry 0: got %q, want %q", entries[0].prefix, "github.com/pinecone-io/")
	}
	if entries[1].prefix != "github.com/foo/bar" {
		t.Errorf("entry 1: got %q, want %q", entries[1].prefix, "github.com/foo/bar")
	}
	if entries[2].prefix != "github.com/baz/qux" {
		t.Errorf("entry 2: got %q, want %q", entries[2].prefix, "github.com/baz/qux")
	}
}

func TestLoadDenylistMissing(t *testing.T) {
	_, err := loadDenylist("/nonexistent/path/denylist.txt")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadDenylistEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "denylist.txt")
	content := `# only comments

# another comment

`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	entries, err := loadDenylist(path)
	if err != nil {
		t.Fatalf("loadDenylist returned error: %v", err)
	}

	if len(entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(entries))
	}
}

func testGraph() *coreDepGraph {
	return &coreDepGraph{
		packages: map[string]depGraphPkg{
			"github.com/lith-project/lith/internal/core/engine": {
				importPath: "github.com/lith-project/lith/internal/core/engine",
				modulePath: "github.com/lith-project/lith",
				imports:    []string{"github.com/lith-project/lith/internal/core/helper"},
			},
			"github.com/lith-project/lith/internal/core/helper": {
				importPath: "github.com/lith-project/lith/internal/core/helper",
				modulePath: "github.com/lith-project/lith",
				imports:    []string{"github.com/pinecone-io/go-pinecone/client"},
			},
			"github.com/pinecone-io/go-pinecone/client": {
				importPath: "github.com/pinecone-io/go-pinecone/client",
				modulePath: "github.com/pinecone-io/go-pinecone",
				imports:    []string{},
			},
		},
	}
}

func TestCheckDenylistClean(t *testing.T) {
	graph := &coreDepGraph{
		packages: map[string]depGraphPkg{
			"github.com/lith-project/lith/internal/core/engine": {
				importPath: "github.com/lith-project/lith/internal/core/engine",
				modulePath: "github.com/lith-project/lith",
				imports:    []string{"github.com/lith-project/lith/internal/core/helper"},
			},
			"github.com/lith-project/lith/internal/core/helper": {
				importPath: "github.com/lith-project/lith/internal/core/helper",
				modulePath: "github.com/lith-project/lith",
				imports:    []string{"fmt"},
			},
		},
	}

	denylist := []denylistEntry{
		{prefix: "github.com/pinecone-io/"},
	}

	violations := checkDenylist(graph, denylist)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(violations))
		for _, v := range violations {
			t.Logf("  %s: %s %s", v.Assertion, v.Package, v.Detail)
		}
	}
}

func TestCheckDenylistViolation(t *testing.T) {
	graph := testGraph()

	denylist := []denylistEntry{
		{prefix: "github.com/pinecone-io/"},
	}

	violations := checkDenylist(graph, denylist)
	// Both engine and helper are core packages that transitively import the denied module
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}

	for _, v := range violations {
		if v.Assertion != "C-4" {
			t.Errorf("assertion: got %q, want %q", v.Assertion, "C-4")
		}
		if v.Detail == "" {
			t.Error("detail should not be empty")
		}
	}
}

func TestCheckDenylistPrefixNotSubstring(t *testing.T) {
	graph := &coreDepGraph{
		packages: map[string]depGraphPkg{
			"github.com/lith-project/lith/internal/core/engine": {
				importPath: "github.com/lith-project/lith/internal/core/engine",
				modulePath: "github.com/lith-project/lith",
				imports:    []string{"github.com/notpinecone-io/go-pinecone/client"},
			},
			"github.com/notpinecone-io/go-pinecone/client": {
				importPath: "github.com/notpinecone-io/go-pinecone/client",
				modulePath: "github.com/notpinecone-io/go-pinecone",
				imports:    []string{},
			},
		},
	}

	denylist := []denylistEntry{
		{prefix: "github.com/pinecone-io/"},
	}

	violations := checkDenylist(graph, denylist)
	if len(violations) != 0 {
		t.Errorf("expected 0 violations (prefix not substring), got %d", len(violations))
		for _, v := range violations {
			t.Logf("  %s: %s %s", v.Assertion, v.Package, v.Detail)
		}
	}
}

func TestCheckDenylistPrefixMatches(t *testing.T) {
	graph := &coreDepGraph{
		packages: map[string]depGraphPkg{
			"github.com/lith-project/lith/internal/core/engine": {
				importPath: "github.com/lith-project/lith/internal/core/engine",
				modulePath: "github.com/lith-project/lith",
				imports:    []string{"github.com/pinecone-io/go-pinecone/client"},
			},
			"github.com/pinecone-io/go-pinecone/client": {
				importPath: "github.com/pinecone-io/go-pinecone/client",
				modulePath: "github.com/pinecone-io/go-pinecone",
				imports:    []string{},
			},
		},
	}

	denylist := []denylistEntry{
		{prefix: "github.com/pinecone-io/"},
	}

	violations := checkDenylist(graph, denylist)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Assertion != "C-4" {
		t.Errorf("assertion: got %q, want %q", violations[0].Assertion, "C-4")
	}
}

func TestTraceChain(t *testing.T) {
	graph := &coreDepGraph{
		packages: map[string]depGraphPkg{
			"A": {importPath: "A", modulePath: "mod/a", imports: []string{"B"}},
			"B": {importPath: "B", modulePath: "mod/b", imports: []string{"C"}},
			"C": {importPath: "C", modulePath: "mod/c", imports: []string{"D"}},
			"D": {importPath: "D", modulePath: "denied.com/x", imports: []string{}},
		},
	}

	chain := traceChain(graph, "A", "denied.com/")
	if chain == nil {
		t.Fatal("expected non-nil chain")
	}
	if len(chain) != 4 {
		t.Fatalf("expected chain of length 4, got %d", len(chain))
	}
	expected := []string{"A", "B", "C", "D"}
	for i, want := range expected {
		if chain[i] != want {
			t.Errorf("chain[%d]: got %q, want %q", i, chain[i], want)
		}
	}
}

func TestTraceChainNoPath(t *testing.T) {
	graph := &coreDepGraph{
		packages: map[string]depGraphPkg{
			"A": {importPath: "A", modulePath: "mod/a", imports: []string{"B"}},
			"B": {importPath: "B", modulePath: "mod/b", imports: []string{}},
		},
	}

	chain := traceChain(graph, "A", "denied.com/")
	if chain != nil {
		t.Errorf("expected nil chain, got %v", chain)
	}
}

func TestBuildCoreDepGraph(t *testing.T) {
	graph, err := buildCoreDepGraph("github.com/lith-project/lith")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(graph.packages) == 0 {
		t.Error("expected core dependency graph to be non-empty once internal/core/ has packages")
	}
	// Every resolved package that lives in the main module must be under
	// internal/core/ (the graph is scoped to ./internal/core/...).
	for path := range graph.packages {
		if strings.HasPrefix(path, "github.com/lith-project/lith/") && !isCorePackage(path) {
			t.Errorf("graph contains non-core main-module package %q", path)
		}
	}
}
