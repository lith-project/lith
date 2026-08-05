package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

type goListPackage struct {
	ImportPath string   `json:"ImportPath"`
	Imports    []string `json:"Imports"`
}

func main() {
	modulePrefix := "github.com/lith-project/lith"

	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: creating stdout pipe: %v\n", err)
		os.Exit(2)
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error: go list failed: %v\n", err)
		os.Exit(2)
	}

	var pkgs []PackageInfo
	dec := json.NewDecoder(stdout)
	for dec.More() {
		var p goListPackage
		if err := dec.Decode(&p); err != nil {
			fmt.Fprintf(os.Stderr, "error: decoding go list output: %v\n", err)
			os.Exit(2)
		}
		pkgs = append(pkgs, PackageInfo{
			ImportPath: p.ImportPath,
			Imports:    p.Imports,
		})
	}

	if err := cmd.Wait(); err != nil {
		fmt.Fprintf(os.Stderr, "error: go list failed: %v\n", err)
		os.Exit(2)
	}

	violations := CheckRules(pkgs, modulePrefix)
	for _, v := range violations {
		fmt.Fprintf(os.Stderr, "%s: %s %s\n", v.Assertion, v.Package, v.Detail)
	}

	if len(violations) > 0 {
		os.Exit(1)
	}
}
