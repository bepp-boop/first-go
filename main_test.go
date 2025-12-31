package main

import (
	"fmt"
	"os"
	"path/filepath"
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
		{"inputs.name", "process.env.INPUTS_NAME"},
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
	yamlfile := filepath.Join("vuln", "ADES100.yaml")
	originalYAML, err := os.ReadFile(yamlfile)
	if err != nil {
		t.Fatalf("error reading YAML file: %v", err)
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
	t.Log(modified)
}

func TestADES101Fix(t *testing.T) {
	yamlfile := filepath.Join("vuln", "ADES101.yaml")
	originalYAML, err := os.ReadFile(yamlfile)
	if err != nil {
		t.Fatalf("error reading YAML file: %v", err)
	}

	// Create the fix (this does NOT apply it yet)
	fix := autofix.ADES101Fix(
		"${{ inputs.name }}",
		"example_job",
		0,
		`console.log('Hello ${{ inputs.name }}')`,
	)
	// Apply the fix
	modified, changed, err := fix.Apply(string(originalYAML))
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
	t.Log(modified)
}

func TestADES102Fix(t *testing.T) {
	yamlfile := filepath.Join("vuln", "ADES102.yaml")
	originalYAML, err := os.ReadFile(yamlfile)
	if err != nil {
		t.Fatalf("error reading YAML file: %v", err)
	}
	// Create the fix (this does NOT apply it yet)
	fix := autofix.ADES102Fix(
		"${{ github.event.issue.title }}",
		"example_job",
		0,
		"Closing ${{ github.event.issue.title }}",
	)
	// Apply the fix
	modified, changed, err := fix.Apply(string(originalYAML))
	if err != nil {
		t.Fatalf("unexpected error applying fix: %v", err)
		t.Log(modified)
	}
	if !changed {
		t.Fatalf("expected fix to report changes, but changed=false")
	}
	// Assert: environment variable is added
	if !strings.Contains(modified, "env:") {
		t.Errorf("expected env block to be added")
	}
	if !strings.Contains(modified, "GITHUB_EVENT_ISSUE_TITLE") {
		t.Errorf("expected generated environment variable name GITHUB_EVENT_ISSUE_TITLE")
	}
	// Assert: script uses process.env
	if !strings.Contains(modified, "process.env.GITHUB_EVENT_ISSUE_TITLE") {
		t.Errorf("expected script to reference process.env.GITHUB_EVENT_ISSUE_TITLE")
	}
	t.Log(modified)
}

func TestADES103Fix(t *testing.T) {
	yamlfile := filepath.Join("vuln", "ADES103.yaml")
	originalYAML, err := os.ReadFile(yamlfile)
	if err != nil {
		t.Fatalf("error reading YAML file: %v", err)
	}
	// Create the fix (this does NOT apply it yet)
	fix := autofix.ADES103Fix(
		"${{ github.event.issue.title }}",
		"example_job",
		0,
		"Closing ${{ github.event.issue.title }}",
	)
	// Apply the fix
	modified, changed, err := fix.Apply(string(originalYAML))
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
	if !strings.Contains(modified, "GITHUB_EVENT_ISSUE_TITLE") {
		t.Errorf("expected generated environment variable name GITHUB_EVENT_ISSUE_TITLE")
	}
	// Assert: script uses shell variable
	if !strings.Contains(modified, "${process.env.GITHUB_EVENT_ISSUE_TITLE}") {
		t.Errorf("expected script to reference $GITHUB_EVENT_ISSUE_TITLE")
	}
	t.Log(modified)
}

func TestADES104Fix(t *testing.T) {
	yamlfile := filepath.Join("vuln", "ADES104.yaml")
	originalYAML, err := os.ReadFile(yamlfile)
	if err != nil {
		t.Fatalf("error reading YAML file: %v", err)
	}
	// Create the fix (this does NOT apply it yet)
	fix := autofix.ADES104Fix(
		"${{ github.event.inputs.file }}",
		"example_job",
		0,
		"jq .version ${{ github.event.inputs.file }} -r",
	)
	// Apply the fix
	modified, changed, err := fix.Apply(string(originalYAML))
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
	if !strings.Contains(modified, "GITHUB_EVENT_INPUTS_FILE") {
		t.Errorf("expected generated environment variable name GITHUB_EVENT_INPUTS_FILE")
	}
	// Assert: script uses shell variable
	if !strings.Contains(modified, "$GITHUB_EVENT_INPUTS_FILE") {
		t.Errorf("expected script to reference $GITHUB_EVENT_INPUTS_FILE")
	}
	t.Log(modified)
}

func TestADES105Fix(t *testing.T) {
	yamlfile := filepath.Join("vuln", "ADES105.yaml")
	originalYAML, err := os.ReadFile(yamlfile)
	if err != nil {
		t.Fatalf("error reading YAML file: %v", err)
	}
	// Create the fix (this does NOT apply it yet)
	fix := autofix.ADES105Fix(
		"${{ inputs.cmd }}",
		"example_job",
		0,
		`echo "Running command: ${{ inputs.cmd }}"`,
	)
	// Apply the fix
	modified, changed, err := fix.Apply(string(originalYAML))
	if err != nil {
		t.Fatalf("unexpected error applying fix: %v", err)
	}
	if !changed {
		t.Fatalf("expected fix to report changes, but changed=false")
	}
	if strings.Contains(modified, "${{ inputs.cmd }}") {
		t.Errorf("No clarify expression should remain in the modified YAML")
	}
	t.Log(modified)
}

func TestADES106Fix(t *testing.T) {
	yamlfile := filepath.Join("vuln", "ADES106.yaml")
	originalYAML, err := os.ReadFile(yamlfile)
	if err != nil {
		t.Fatalf("error reading YAML file: %v", err)
	}
	// Create the fix (this does NOT apply it yet)
	fix := autofix.ADES106Fix(
		"${{ inputs.value }}",
		"example_job",
		0,
		"1 + parseInt(${{ inputs.value }})",
	)
	// Apply the fix
	modified, changed, err := fix.Apply(string(originalYAML))
	if err != nil {
		t.Fatalf("unexpected error applying fix: %v", err)
	}
	if !changed {
		t.Fatalf("expected fix to report changes, but changed=false")
	}
	if !strings.Contains(modified, "env.INPUTS_VALUE") {
		t.Errorf("expected to have this expression format")
	}
	t.Log(modified)
}

// Need to fix remove ”
func TestADES107Fix(t *testing.T) {
	yamlpatch := filepath.Join("vuln", "ADES107.yaml")
	originalYAML, err := os.ReadFile(yamlpatch)
	if err != nil {
		t.Fatalf("error reading YAML file: %v", err)
	}

	// Create the fix
	fix := autofix.ADES107Fix(
		"${{ inputs.color }}",
		"example_job",
		0,
		"{ attachments: [{ color: '${{ inputs.color }}' }] }",
	)
	// Apply the fix
	modified, changed, err := fix.Apply(string(originalYAML))
	if err != nil {
		t.Fatalf("unexpected error applying fix: %v", err)
	}
	if !changed {
		t.Fatalf("expected fix to report changes, but changed=false")
	}
	if strings.Contains(modified, "''${process.env.INPUTS_COLOR}''") {
		t.Fatalf("Still have the single quote around it, format issues")
		fmt.Println(modified)
	}
	t.Log(modified)
}

// Change from ” to "" for interpolation
func TestADES108Fix(t *testing.T) {
	yamlpatch := filepath.Join("vuln", "ADES108.yaml")
	originalYAML, err := os.ReadFile(yamlpatch)
	if err != nil {
		t.Fatalf("error reading YAML file: %v", err)
	}

	fix := autofix.ADES108Fix(
		"${{ inputs.name }}",
		"example_job",
		0,
		"echo 'Hello ${{ inputs.name }}'",
	)

	// Apply the fix
	modified, changed, err := fix.Apply(string(originalYAML))
	if err != nil {
		t.Fatalf("unexpected error applying fix: %v", err)
	}
	if !changed {
		t.Fatalf("expected fix to report changes, but changed=false")
	}
	if strings.Contains(modified, "''${process.env.INPUTS_COLOR}''") {
		t.Fatalf("Still have the single quote around it, format issues")
	}
	t.Log(modified)
}

// Need to fix for python interpolation because of double quotation
func TestADES109Fix(t *testing.T) {
	yamlpatch := filepath.Join("vuln", "ADES109.yaml")
	originalYAML, err := os.ReadFile(yamlpatch)
	if err != nil {
		t.Fatalf("error reading YAML file: %v", err)
	}

	fix := autofix.ADES109Fix(
		"${{ inputs.name }}",
		"example_job",
		0,
		"print(\"Hello ${{ inputs.name }}\")",
	)
	// Apply the fix
	modified, changed, err := fix.Apply(string(originalYAML))
	if err != nil {
		t.Fatalf("unexpected error applying fix: %v", err)
	}
	if !changed {
		t.Fatalf("expected fix to report changes, but changed=false")
	}
	if strings.Contains(modified, "''${process.env.INPUTS_COLOR}''") {
		t.Fatalf("Still have the single quote around it, format issues")
	}
	t.Log(modified)

}

func TestADES110Fix(t *testing.T) {
	yamlpatch := filepath.Join("vuln", "ADES110.yaml")
	originalYAML, err := os.ReadFile(yamlpatch)
	if err != nil {
		t.Fatalf("error reading YAML file: %v", err)
	}

	fix := autofix.ADES110Fix(
		"${{ inputs.name }}",
		"example_job",
		0,
		"Write-Output 'Hello ${{ inputs.name }}'",
	)
	// Apply the fix
	modified, changed, err := fix.Apply(string(originalYAML))
	if err != nil {
		t.Fatalf("unexpected error applying fix: %v", err)
	}
	if !changed {
		t.Fatalf("expected fix to report changes, but changed=false")
	}
	if strings.Contains(modified, "''${process.env.INPUTS_COLOR}''") {
		t.Fatalf("Still have the single quote around it, format issues")
	}
	t.Log(modified)

}

func TestADES111Fix(t *testing.T) {
	yamlpatch := filepath.Join("vuln", "ADES111.yaml")
	originalYAML, err := os.ReadFile(yamlpatch)
	if err != nil {
		t.Fatalf("error reading YAML file: %v", err)
	}

	fix := autofix.ADES111Fix(
		"${{ inputs.query }}",
		"example_job",
		0,
		"yq '${{ inputs.query }}' 'config.yml'",
	)
	// Apply the fix
	modified, changed, err := fix.Apply(string(originalYAML))
	if err != nil {
		t.Fatalf("unexpected error applying fix: %v", err)
	}
	if !changed {
		t.Fatalf("expected fix to report changes, but changed=false")
	}
	if strings.Contains(modified, "''${process.env.INPUTS_COLOR}''") {
		t.Fatalf("Still have the single quote around it, format issues")
	}
	t.Log(modified)

}

func TestADES112Fix(t *testing.T) {
	yamlpatch := filepath.Join("vuln", "ADES112.yaml")
	originalYAML, err := os.ReadFile(yamlpatch)
	if err != nil {
		t.Fatalf("error reading YAML file: %v", err)
	}

	fix := autofix.ADES112Fix(
		"${{ inputs.query }}",
		"example_job",
		0,
		"yq eval '${{ inputs.query }}' 'config.yml'",
	)
	// Apply the fix
	modified, changed, err := fix.Apply(string(originalYAML))
	if err != nil {
		t.Fatalf("unexpected error applying fix: %v", err)
	}
	if !changed {
		t.Fatalf("expected fix to report changes, but changed=false")
	}
	if strings.Contains(modified, "''${process.env.INPUTS_COLOR}''") {
		t.Fatalf("Still have the single quote around it, format issues")
	}
	t.Log(modified)

}

func TestADES113Fix(t *testing.T) {
	yamlpatch := filepath.Join("vuln", "ADES113.yaml")
	originalYAML, err := os.ReadFile(yamlpatch)
	if err != nil {
		t.Fatalf("error reading YAML file: %v", err)
	}

	fix := autofix.ADES113Fix(
		"${{ inputs.name }}",
		"example_job",
		0,
		"Write-Output 'Hello ${{ inputs.name }}'",
	)
	// Apply the fix
	modified, changed, err := fix.Apply(string(originalYAML))
	if err != nil {
		t.Fatalf("unexpected error applying fix: %v", err)
	}
	if !changed {
		t.Fatalf("expected fix to report changes, but changed=false")
	}
	if strings.Contains(modified, "''${process.env.INPUTS_COLOR}''") {
		t.Fatalf("Still have the single quote around it, format issues")
	}
	t.Log(modified)

}
