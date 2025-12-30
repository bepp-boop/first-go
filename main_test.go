package main

import (
	"fmt"
	"strings"
	"testing"

	"example.com/first-go/autofix"
)

// Test function
func TestGenerateEnvVarName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"github.event.issue.title", "GITHUB_EVENT_ISSUE_TITLE"},
		{"secret.api-config", "SECRET_API_CONFIG"},
		{"fromJSON(secrets.config)", "CONFIG"},
		{"fromJSON(secrets.database-config).username", "DATABASE_CONFIG_USERNAME"},
		{"fromJSON(secrets.api-config)", "API_CONFIG"},
		{"fromJSON(secrets.database_config).password", "DATABASE_CONFIG_PASSWORD"},
		{"github.event@issue.title", "GITHUB_EVENT_ISSUE_TITLE"},
		{"my var.name-test@prod", "MY_VAR_NAME_TEST_PROD"},
		{"config.v2.test", "CONFIG_V2_TEST"},
		{"inputs.name", "INPUTS_NAME"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := autofix.GenerateEnvVarName(tt.input)
			if got != tt.expected {
				t.Errorf("generateEnvVarName(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGenerateProcessedEnvVarName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"github.event.issue.title", "process.env.GITHUB_EVENT_ISSUE_TITLE"},
		{"inputs.name", "process.env.NAME"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := autofix.GenerateProcessedEnvVarName(tt.input)
			if got != tt.expected {
				t.Errorf("GenerateProcessedEnvVarName(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestADES100Fix(t *testing.T) {
	originalYAML := `
name: test

on: workflow_dispatch

jobs:
  test-job:
    runs-on: ubuntu-latest
    steps:
      - name: Example step
        run: echo "Hello ${{ inputs.name }}"
`

	// Create the fix (this does NOT apply it yet)
	fix := autofix.ADES100Fix(
		"${{ inputs.name }}",
		"test-job",
		0,
		`echo "Hello ${{ inputs.name }}"`,
	)

	// Apply the fix
	modified, changed, err := fix.Apply(originalYAML)
	if err != nil {
		t.Fatalf("unexpected error applying fix: %v", err)
	}

	if !changed {
		t.Fatalf("expected fix to report changes, but changed=false")
	}

	// Assert: environment variable is added
	if !strings.Contains(modified, "env:") {
		t.Errorf("expected env block to be added")
	}

	if !strings.Contains(modified, "INPUTS_NAME") {
		t.Errorf("expected generated environment variable name INPUTS_NAME")
	}

	// Assert: run script uses shell variable
	if !strings.Contains(modified, "$INPUTS_NAME") {
		t.Errorf("expected run script to reference $INPUTS_NAME")
	}
	fmt.Println(modified)
}

func TestADES101Fix(t *testing.T) {
	originalYAML := `
name: ADES101

on: workflow_dispatch

jobs:
  example_job:
    runs-on: ubuntu-latest
    steps:
      - name: Example step
        uses: actions/github-script@v6
        with:
          script: console.log('Hello ${{ inputs.name }}')
`
	// Create the fix (this does NOT apply it yet)
	fix := autofix.ADES101Fix(
		"${{ inputs.name }}",
		"example_job",
		0,
		`console.log('Hello ${{ inputs.name }}')`,
	)
	// Apply the fix
	modified, changed, err := fix.Apply(originalYAML)
	if err != nil {
		t.Fatalf("unexpected error applying fix: %v", err)
	}
	if !changed {
		t.Fatalf("expected fix to report changes, but changed=false")
	}
	// Assert: environment variable is added
	if !strings.Contains(modified, "env:") {
		t.Errorf("expected env block to be added")
	}

	if !strings.Contains(modified, "INPUTS_NAME") {
		t.Errorf("expected generated environment variable name INPUTS_NAME")
	}
	// Assert: script uses process.env
	if !strings.Contains(modified, "process.env.INPUTS_NAME") {
		t.Errorf("expected script to reference process.env.INPUTS_NAME")
	}
	fmt.Println(modified)
}
