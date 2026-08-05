package main

import "testing"

func TestClassify(t *testing.T) {
	module := "github.com/lith-project/lith"
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"cmd package", module + "/cmd/lithd", "cmd"},
		{"cmd root", module + "/cmd", "cmd"},
		{"core package", module + "/internal/core/config", "core"},
		{"core root", module + "/internal/core", "core"},
		{"adapter package", module + "/internal/adapter/cli", "adapter"},
		{"adapter root", module + "/internal/adapter", "adapter"},
		{"plugin package", module + "/internal/plugin/mcp", "plugin"},
		{"plugin root", module + "/internal/plugin", "plugin"},
		{"tools package", module + "/tools/conformance/boundary", ""},
		{"external package", "github.com/foo/bar", ""},
		{"stdlib package", "fmt", ""},
		{"unclassified", module + "/docs/foo", ""},
		{"module root", module, ""},
		{"similar prefix", "github.com/lith-project/lithfoo/cmd/x", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classify(tt.path, module)
			if got != tt.expected {
				t.Errorf("classify(%q, %q) = %q, want %q", tt.path, module, got, tt.expected)
			}
		})
	}
}

func TestCheckRulesClean(t *testing.T) {
	module := "github.com/lith-project/lith"
	pkgs := []PackageInfo{
		{ImportPath: module + "/cmd/lithd", Imports: []string{module + "/internal/core/config"}},
		{ImportPath: module + "/internal/core/config", Imports: []string{"fmt", "os"}},
	}
	violations := CheckRules(pkgs, module)
	if len(violations) != 0 {
		t.Errorf("expected no violations, got %d", len(violations))
		for _, v := range violations {
			t.Logf("  %s: %s %s", v.Assertion, v.Package, v.Detail)
		}
	}
}

func TestCheckRulesToolsExempt(t *testing.T) {
	module := "github.com/lith-project/lith"
	pkgs := []PackageInfo{
		{ImportPath: module + "/tools/conformance/boundary", Imports: []string{"fmt"}},
	}
	violations := CheckRules(pkgs, module)
	if len(violations) != 0 {
		t.Errorf("tools/ package should be exempt, got %d violations", len(violations))
	}
}

func TestCheckRulesC1(t *testing.T) {
	module := "github.com/lith-project/lith"
	pkgs := []PackageInfo{
		{ImportPath: module + "/docs/foo", Imports: []string{}},
	}
	violations := CheckRules(pkgs, module)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Assertion != "C-1" {
		t.Errorf("expected C-1, got %s", violations[0].Assertion)
	}
}

func TestCheckRulesC2(t *testing.T) {
	module := "github.com/lith-project/lith"
	pkgs := []PackageInfo{
		{ImportPath: module + "/internal/core/engine", Imports: []string{module + "/internal/adapter/cli"}},
	}
	violations := CheckRules(pkgs, module)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Assertion != "C-2" {
		t.Errorf("expected C-2, got %s", violations[0].Assertion)
	}
}

func TestCheckRulesC2Plugin(t *testing.T) {
	module := "github.com/lith-project/lith"
	pkgs := []PackageInfo{
		{ImportPath: module + "/internal/core/engine", Imports: []string{module + "/internal/plugin/mcp"}},
	}
	violations := CheckRules(pkgs, module)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Assertion != "C-2" {
		t.Errorf("expected C-2, got %s", violations[0].Assertion)
	}
}

func TestCheckRulesC2Cmd(t *testing.T) {
	module := "github.com/lith-project/lith"
	pkgs := []PackageInfo{
		{ImportPath: module + "/internal/core/engine", Imports: []string{module + "/cmd/lithd"}},
	}
	violations := CheckRules(pkgs, module)
	// C-2: core imports cmd; C-5: any package imports cmd
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(violations))
	}
	assertions := make(map[string]int)
	for _, v := range violations {
		assertions[v.Assertion]++
	}
	if assertions["C-2"] != 1 {
		t.Errorf("expected 1 C-2 violation, got %d", assertions["C-2"])
	}
	if assertions["C-5"] != 1 {
		t.Errorf("expected 1 C-5 violation, got %d", assertions["C-5"])
	}
}

func TestCheckRulesC5(t *testing.T) {
	module := "github.com/lith-project/lith"
	pkgs := []PackageInfo{
		{ImportPath: module + "/internal/adapter/cli", Imports: []string{module + "/cmd/lithd"}},
	}
	violations := CheckRules(pkgs, module)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Assertion != "C-5" {
		t.Errorf("expected C-5, got %s", violations[0].Assertion)
	}
}

func TestCheckRulesMultipleViolations(t *testing.T) {
	module := "github.com/lith-project/lith"
	pkgs := []PackageInfo{
		{
			ImportPath: module + "/internal/core/engine",
			Imports:    []string{module + "/internal/adapter/cli", module + "/cmd/lithd"},
		},
		{ImportPath: module + "/docs/foo", Imports: []string{}},
	}
	violations := CheckRules(pkgs, module)
	// core/engine: C-2 (imports adapter), C-2 (imports cmd), C-5 (imports cmd)
	// docs/foo: C-1 (unclassified)
	if len(violations) != 4 {
		t.Fatalf("expected 4 violations, got %d", len(violations))
	}
	assertions := make(map[string]int)
	for _, v := range violations {
		assertions[v.Assertion]++
	}
	if assertions["C-1"] != 1 {
		t.Errorf("expected 1 C-1 violation, got %d", assertions["C-1"])
	}
	if assertions["C-2"] != 2 {
		t.Errorf("expected 2 C-2 violations, got %d", assertions["C-2"])
	}
	if assertions["C-5"] != 1 {
		t.Errorf("expected 1 C-5 violation, got %d", assertions["C-5"])
	}
}
