package main

import (
	"testing"
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
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := generateEnvVarName(tt.input)
			if got != tt.expected {
				t.Errorf("generateEnvVarName(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
