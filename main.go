package main

import (
	"fmt"
	"regexp"
	"strings"

	yamlpatch "github.com/palantir/pkg/yamlpatch"
)

func ApplyYamlReplace(original []byte, path string, value interface{}) ([]byte, error) {
	patch := yamlpatch.Patch{
		yamlpatch.Operation{
			Type:  yamlpatch.OperationReplace,
			Path:  yamlpatch.MustParsePath(path),
			Value: value,
		},
	}

	return yamlpatch.Apply(original, patch)
}

func create_env_var_fix(expression string, job_id string, step_index, script string) {

}

func generateEnvVarName(expr string) string {
	expr = strings.TrimSpace(expr)
	cleaned := expr

	// Handle fromJSON(secrets.*) patterns
	if strings.HasPrefix(expr, "fromJSON(secrets.") {
		closing := strings.Index(expr, ")")
		if closing != -1 {
			secretsPart := expr[17:closing] // skip "fromJSON(secrets."
			propertyPart := expr[closing+1:]

			if strings.HasPrefix(propertyPart, ".") {
				cleaned = fmt.Sprintf("%s_%s", secretsPart, propertyPart[1:])
			} else if propertyPart == "" {
				cleaned = secretsPart
			} else {
				cleaned = fmt.Sprintf("%s_%s", secretsPart, propertyPart)
			}
		}
	} else if strings.HasPrefix(expr, "fromJSON(") {
		closing := strings.Index(expr, ")")
		if closing != -1 {
			inner := expr[9:closing] // skip "fromJSON("
			propertyPart := expr[closing+1:]

			if strings.HasPrefix(propertyPart, ".") {
				cleaned = fmt.Sprintf("%s_%s", inner, propertyPart[1:])
			} else if propertyPart == "" {
				cleaned = inner
			} else {
				cleaned = fmt.Sprintf("%s_%s", inner, propertyPart)
			}
		}
	} else if strings.HasPrefix(expr, "secrets.") {
		cleaned = strings.TrimPrefix(expr, "secrets.")
	}

	// Replace non-alphanumeric characters with _
	reg := regexp.MustCompile(`[^a-zA-Z0-9]`)
	cleaned = reg.ReplaceAllString(cleaned, "_")

	// Convert to uppercase
	cleaned = strings.ToUpper(cleaned)

	return cleaned
}

func main() {
	original := []byte(`
- name: Example step 
  run: |
    title="${{ github.event.issue.title }}"
    echo "Issue title: $title"`)

	patch_edit := yamlpatch.Patch{
		yamlpatch.Operation{
			Type:  yamlpatch.OperationReplace,
			Path:  yamlpatch.MustParsePath("/"),
			Value: "Change Name",
		},
	}

	// Apply the patch
	updated, err := yamlpatch.Apply(original, patch_edit)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(updated))
	fmt.Println(generateEnvVarName("github.event.issue.title")) // GITHUB_EVENT_ISSUE_TITLE

}
