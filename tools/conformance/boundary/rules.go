package main

import "strings"

// PackageInfo holds the information about a Go package from go list.
type PackageInfo struct {
	ImportPath string
	Imports    []string
}

// Violation represents a boundary rule violation.
type Violation struct {
	Assertion string
	Package   string
	Detail    string
}

// classify returns the architectural layer for the given import path.
// Returns "cmd", "core", "adapter", "plugin", or "" for unclassified.
func classify(importPath, modulePrefix string) string {
	if importPath == modulePrefix {
		return ""
	}
	prefix := modulePrefix + "/"
	if !strings.HasPrefix(importPath, prefix) {
		return ""
	}
	rel := importPath[len(prefix):]

	if rel == "cmd" || strings.HasPrefix(rel, "cmd/") {
		return "cmd"
	}
	if rel == "internal/core" || strings.HasPrefix(rel, "internal/core/") {
		return "core"
	}
	if rel == "internal/adapter" || strings.HasPrefix(rel, "internal/adapter/") {
		return "adapter"
	}
	if rel == "internal/plugin" || strings.HasPrefix(rel, "internal/plugin/") {
		return "plugin"
	}
	return ""
}

// CheckRules checks all boundary rules against the given packages.
func CheckRules(pkgs []PackageInfo, modulePrefix string) []Violation {
	var violations []Violation

	for _, pkg := range pkgs {
		if isToolsPackage(pkg.ImportPath, modulePrefix) {
			continue
		}

		pkgClass := classify(pkg.ImportPath, modulePrefix)

		// C-1: Package outside tools/ must be classified into a known layer
		if pkgClass == "" {
			violations = append(violations, Violation{
				Assertion: "C-1",
				Package:   pkg.ImportPath,
				Detail:    "unclassified package (not in cmd/, internal/core/, internal/adapter/, or internal/plugin/)",
			})
		}

		// C-2: core packages must not import adapter, plugin, or cmd
		if pkgClass == "core" {
			for _, imp := range pkg.Imports {
				impClass := classify(imp, modulePrefix)
				if impClass == "adapter" || impClass == "plugin" || impClass == "cmd" {
					violations = append(violations, Violation{
						Assertion: "C-2",
						Package:   pkg.ImportPath,
						Detail:    "core package imports " + impClass + " package " + imp,
					})
				}
			}
		}

		// C-5: No package may import cmd
		for _, imp := range pkg.Imports {
			impClass := classify(imp, modulePrefix)
			if impClass == "cmd" {
				violations = append(violations, Violation{
					Assertion: "C-5",
					Package:   pkg.ImportPath,
					Detail:    "package imports cmd package " + imp,
				})
			}
		}
	}

	return violations
}

func isToolsPackage(importPath, modulePrefix string) bool {
	prefix := modulePrefix + "/tools"
	return importPath == prefix || strings.HasPrefix(importPath, prefix+"/")
}
