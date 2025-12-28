package autofix

import (
	"fmt"
	"regexp"
	"strings"

	yamlpatch "github.com/palantir/pkg/yamlpatch"
)

func GenerateEnvVarName(expr string) string {
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

func CreateEnvVarFix(expression string, job_id string, step_index int, script string) {
	expressionCopy := expression
	scriptCopy := script

	apply := func(content string) (string, bool, error) {
		var operations []yamlpatch.Operation

		newScript := scriptCopy

		// Check if expression already includes ${{ }}
		exprTrim := strings.TrimSpace(expressionCopy)
		var cleanExpr, fullExpr string

		if strings.HasPrefix(exprTrim, "${{") && strings.HasSuffix(exprTrim, "}}") {
			cleanExpr = strings.TrimSpace(exprTrim[3 : len(exprTrim)-2])
			fullExpr = exprTrim
		} else {
			cleanExpr = exprTrim
			fullExpr = fmt.Sprintf("${{ %s }}", cleanExpr)
		}

		// Generate environment variable name
		envVarName := GenerateEnvVarName(cleanExpr)

		// Replace expression in script
		if !strings.Contains(newScript, fullExpr) {
			// Equivalent to Ok(None) in Rust
			return content, false, nil
		}

		// Replace the expression with an env var reference in the script
		newScript = strings.ReplaceAll(
			newScript,
			fullExpr,
			fmt.Sprintf("$%s", envVarName),
		)
		//Print the new script
		fmt.Println("Modified Script:\n", newScript)

		// Replace run script
		runPath := fmt.Sprintf("/jobs/%s/steps/%d", job_id, step_index)
		operations = append(operations, yamlpatch.Operation{
			Type:  yamlpatch.OperationReplace,
			Path:  yamlpatch.MustParsePath(runPath),
			Value: newScript,
		})

		// Merge env block into step
		envPath := fmt.Sprintf("/jobs/%s/steps/%d", job_id, step_index)
		operations = append(operations, yamlpatch.Operation{
			Type: yamlpatch.OperationAdd,
			Path: yamlpatch.MustParsePath(envPath),
			Value: map[string]interface{}{
				"env": map[string]interface{}{
					envVarName: fmt.Sprintf("${{ %s }}", cleanExpr),
				},
			},
		})

		// Apply YAML patch
		modified, err := yamlpatch.Apply([]byte(content), yamlpatch.Patch(operations))
		if err != nil {
			return "", false, err
		}

		// Equivalent to Ok(Some(new_content))
		return string(modified), true, nil
	}

	modified, changed, err := apply(scriptCopy)
	if err != nil {
		fmt.Println("Error applying fix:", err)
		return
	}

	if changed {
		fmt.Println("Fix applied successfully.")
		fmt.Println(modified)
	} else {
		fmt.Println("No changes made.")
	}
}

func CreateCompositeEnvVarFix(expression string, step_index int, script string) {
	expressionCopy := expression
	scriptCopy := script

	apply := func(content string) (string, bool, error) {
		var operations []yamlpatch.Operation

		newScript := scriptCopy

		// Check if expression already includes ${{ }}
		exprTrim := strings.TrimSpace(expressionCopy)
		var cleanExpr, fullExpr string

		if strings.HasPrefix(exprTrim, "${{") && strings.HasSuffix(exprTrim, "}}") {
			cleanExpr = strings.TrimSpace(exprTrim[3 : len(exprTrim)-2])
			fullExpr = exprTrim
		} else {
			cleanExpr = exprTrim
			fullExpr = fmt.Sprintf("${{ %s }}", cleanExpr)
		}

		// Generate environment variable name
		envVarName := GenerateEnvVarName(cleanExpr)

		// Replace expression in script
		if !strings.Contains(newScript, fullExpr) {
			// Equivalent to Ok(None) in Rust
			return content, false, nil
		}

		// Replace the expression with an env var reference in the script
		newScript = strings.ReplaceAll(
			newScript,
			fullExpr,
			fmt.Sprintf("$%s", envVarName),
		)
		//Print the new script
		fmt.Println("Modified Script:\n", newScript)

		// Replace run script
		runPath := fmt.Sprintf("/runs/steps/%d/run", step_index)
		operations = append(operations, yamlpatch.Operation{
			Type:  yamlpatch.OperationReplace,
			Path:  yamlpatch.MustParsePath(runPath),
			Value: newScript,
		})

		// Merge env block into step
		envPath := fmt.Sprintf("/runs/steps/%d", step_index)
		operations = append(operations, yamlpatch.Operation{
			Type: yamlpatch.OperationAdd,
			Path: yamlpatch.MustParsePath(envPath),
			Value: map[string]interface{}{
				"env": map[string]interface{}{
					envVarName: fmt.Sprintf("${{ %s }}", cleanExpr),
				},
			},
		})

		// Apply YAML patch
		modified, err := yamlpatch.Apply([]byte(content), yamlpatch.Patch(operations))
		if err != nil {
			return "", false, err
		}

		// Equivalent to Ok(Some(new_content))
		return string(modified), true, nil
	}

	modified, changed, err := apply(scriptCopy)
	if err != nil {
		fmt.Println("Error applying fix:", err)
		return
	}

	if changed {
		fmt.Println("Fix applied successfully.")
		fmt.Println(modified)
	} else {
		fmt.Println("No changes made.")
	}
}

func AddShellSpecifier(job_id string, step_index int) {
	apply := func(content string) (string, bool, error) {
		var operations []yamlpatch.Operation
		// Add shell: bash to the specified step
		shellPath := fmt.Sprintf("/jobs/%s/steps/%d/shell", job_id, step_index)
		operations = append(operations, yamlpatch.Operation{
			Type:  yamlpatch.OperationAdd,
			Path:  yamlpatch.MustParsePath(shellPath),
			Value: "bash",
		},
		)

	}
}

func AddCompositieShellSpecifier(step_index int) {
	apply := func(content string) (string, bool, error) {
		var operations []yamlpatch.Operation
		// Add shell: bash to the specified step
		shellPath := fmt.Sprintf("/runs/steps/%d/shell", step_index)
		operations = append(operations, yamlpatch.Operation{
			Type:  yamlpatch.OperationAdd,
			Path:  yamlpatch.MustParsePath(shellPath),
			Value: "bash",
		},
		)
	}
}

func Math(a int, b int) int {
	return a + b
}
