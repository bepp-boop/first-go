package main

import (
	"fmt"

	// File handling imports
	"os"
	"path/filepath"
	// Import the autofix package
	// <-- matches module + folder
)

func main() {
	// Import from a different folder yaml and print it in template_injection
	yamlfile := filepath.Join("vuln", "ADES100.yaml")

	data, err := os.ReadFile(yamlfile)
	if err != nil {
		fmt.Println("Error reading YAML file:", err)
		return
	}

	fmt.Println("YAML file content:\n", string(data))
}
