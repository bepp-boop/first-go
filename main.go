package main

import (
	"fmt"
	"log"

	// File handling imports

	"os"
	"path/filepath"

	// Import the autofix package
	"example.com/first-go/autofix"
	// <-- matches module + folder
)

func main() {
	yamlfile := filepath.Join("vuln", "ADES100.yaml")
	originalYAML, err := os.ReadFile(yamlfile)
	if err != nil {
		log.Fatalf("error reading YAML file: %v", err)
	}

	// Create the fix (this does NOT apply it yet)
	fix := autofix.ADES100Fix(
		"${{ inputs.name }}",
		"test",
		0,
		`echo "Hello ${{ inputs.name }}"`,
	)

	// Apply the fix
	modified, changed, err := fix.Apply(string(originalYAML))
	if err != nil {
		log.Fatalf("error applying fix: %v", err)
	}
	if !changed {
		log.Fatalf("expected the fix to make changes, but it did not")
	}
	fmt.Println("Modified YAML content:")
	fmt.Println(modified)
}
