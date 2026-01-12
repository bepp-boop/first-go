package main

import (
	"fmt"

	// File handling imports
	"flag"
	"os"
	"path/filepath"
	// Import the autofix package
	// <-- matches module + folder
)

type FixApplyFunc func(oldContent string) (newContent *string, err error)

type Fix struct {
	Title       string
	Description string
	Apply       FixApplyFunc
}

func (f Fix) ApplyToContent(oldContent string) (*string, error) {
	return f.Apply(oldContent)
}

func main() {
	// Define flags
	jsonOut := flag.Bool("json", false, "Output result as JSON")
	flag.Parse()

	// After flag.Parse(), remaining args are positional
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: ades [--json] <gha-yaml-file>")
		os.Exit(1)
	}

	// Read YAML path from positional argument
	yamlPath := args[0]

	// Normalize path (important on Windows)
	absPath, err := filepath.Abs(yamlPath)
	if err != nil {
		fmt.Println("Invalid path:", err)
		os.Exit(1)
	}

	fmt.Println("Using YAML:", absPath)
	fmt.Println("JSON output:", *jsonOut)

	// ---- use absPath here ----
	// data, err := os.ReadFile(absPath)
	// ...
}
