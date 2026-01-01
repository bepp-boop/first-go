package main

import (
	"fmt"

	// File handling imports
	"os"
	"path/filepath"

	// Import the autofix package
	// <-- matches module + folder
	yamlpatch "github.com/palantir/pkg/yamlpatch"
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
	// Import from a different folder yaml and print it in template_injection
	yamlfile := filepath.Join("vuln", "ADES107.yaml")

	data, err := os.ReadFile(yamlfile)
	if err != nil {
		fmt.Println("Error reading YAML file:", err)
		return
	}

	test := yamlpatch.Operation{
		Type:  yamlpatch.OperationReplace,
		Path:  yamlpatch.MustParsePath("/jobs/example_job/steps/0/with/custom_payload"),
		Value: "Hello",
	}

	modify, err := yamlpatch.Apply(data, yamlpatch.Patch([]yamlpatch.Operation{test}))
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(modify))

}
