package autofix

import (
	"fmt"
	"regexp"
	"strings"

	yamlpatch "github.com/palantir/pkg/yamlpatch"
)

type Fix struct {
	Title       string
	Description string
	Apply       func(string) (string, bool, error)
}

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

func GenerateProcessedEnvVarName(expr string) string {
	return "process.env." + GenerateEnvVarName(expr)
}

func AddShellSpecifier(jobID string, stepIndex int) Fix {
	return Fix{
		Title:       "Add explicit shell specification",
		Description: "Explicitly specify the shell to ensure predictable execution behavior.",

		Apply: func(content string) (string, bool, error) {
			var operations []yamlpatch.Operation

			stepPath := fmt.Sprintf("/jobs/%s/steps/%d", jobID, stepIndex)
			operations = append(operations, yamlpatch.Operation{
				Type: yamlpatch.OperationAdd,
				Path: yamlpatch.MustParsePath(stepPath),
				Value: map[string]interface{}{
					"shell": "bash",
				},
			})

			modified, err := yamlpatch.Apply([]byte(content), yamlpatch.Patch(operations))
			if err != nil {
				return "", false, err
			}

			return string(modified), true, nil
		},
	}
}

func CreateCompositeShellSpecificationFix(stepIndex int) Fix {
	return Fix{
		Title:       "Add explicit shell specification",
		Description: "Explicitly specify the shell for composite action steps.",

		Apply: func(content string) (string, bool, error) {
			var operations []yamlpatch.Operation

			stepPath := fmt.Sprintf("/runs/steps/%d", stepIndex)
			operations = append(operations, yamlpatch.Operation{
				Type: yamlpatch.OperationAdd,
				Path: yamlpatch.MustParsePath(stepPath),
				Value: map[string]interface{}{
					"shell": "bash",
				},
			})

			modified, err := yamlpatch.Apply([]byte(content), yamlpatch.Patch(operations))
			if err != nil {
				return "", false, err
			}

			return string(modified), true, nil
		},
	}
}

func ADES100Fix(
	expression string,
	jobID string,
	stepIndex int,
	script string,
) Fix {

	expressionCopy := expression
	scriptCopy := script

	return Fix{
		Title: "Move template expression to environment variable",
		Description: fmt.Sprintf(
			"Move template expression (%s) to an environment variable to prevent "+
				"template injection vulnerabilities.",
			expression,
		),

		Apply: func(content string) (string, bool, error) {
			var operations []yamlpatch.Operation
			newScript := scriptCopy

			exprTrim := strings.TrimSpace(expressionCopy)
			var cleanExpr, fullExpr string

			if strings.HasPrefix(exprTrim, "${{") && strings.HasSuffix(exprTrim, "}}") {
				cleanExpr = strings.TrimSpace(exprTrim[3 : len(exprTrim)-2])
				fullExpr = exprTrim
			} else {
				cleanExpr = exprTrim
				fullExpr = fmt.Sprintf("${{ %s }}", cleanExpr)
			}

			envVarName := GenerateEnvVarName(cleanExpr)

			if !strings.Contains(newScript, fullExpr) {
				return content, false, nil
			}

			newScript = strings.ReplaceAll(
				newScript,
				fullExpr,
				fmt.Sprintf("$%s", envVarName),
			)

			runPath := fmt.Sprintf("/jobs/%s/steps/%d/run", jobID, stepIndex)
			operations = append(operations, yamlpatch.Operation{
				Type:  yamlpatch.OperationReplace,
				Path:  yamlpatch.MustParsePath(runPath),
				Value: newScript,
			})

			envPath := fmt.Sprintf("/jobs/%s/steps/%d", jobID, stepIndex)
			operations = append(operations, yamlpatch.Operation{
				Type: yamlpatch.OperationAdd,
				Path: yamlpatch.MustParsePath(envPath),
				Value: map[string]interface{}{
					"env": map[string]interface{}{
						envVarName: fmt.Sprintf("${{ %s }}", cleanExpr),
					},
				},
			})

			modified, err := yamlpatch.Apply([]byte(content), yamlpatch.Patch(operations))
			if err != nil {
				return "", false, err
			}

			return string(modified), true, nil
		},
	}
}
func ADES100FixComposite(
	expression string,
	stepIndex int,
	script string,
) Fix {

	expressionCopy := strings.TrimSpace(expression)
	scriptCopy := script

	return Fix{
		Title: "Move template expression to environment variable",
		Description: fmt.Sprintf(
			"Move template expression (%s) to an environment variable in a composite "+
				"action to prevent template injection vulnerabilities.",
			expression,
		),

		Apply: func(content string) (string, bool, error) {
			var operations []yamlpatch.Operation
			newScript := scriptCopy

			var cleanExpr, fullExpr string
			if strings.HasPrefix(expressionCopy, "${{") && strings.HasSuffix(expressionCopy, "}}") {
				cleanExpr = strings.TrimSpace(expressionCopy[3 : len(expressionCopy)-2])
				fullExpr = expressionCopy
			} else {
				cleanExpr = expressionCopy
				fullExpr = fmt.Sprintf("${{ %s }}", cleanExpr)
			}

			envVarName := GenerateEnvVarName(cleanExpr)

			if !strings.Contains(newScript, fullExpr) {
				return content, false, nil
			}

			newScript = strings.ReplaceAll(
				newScript,
				fullExpr,
				fmt.Sprintf("$%s", envVarName),
			)

			runPath := fmt.Sprintf("/runs/steps/%d/run", stepIndex)
			operations = append(operations, yamlpatch.Operation{
				Type:  yamlpatch.OperationReplace,
				Path:  yamlpatch.MustParsePath(runPath),
				Value: newScript,
			})

			envPath := fmt.Sprintf("/runs/steps/%d", stepIndex)
			operations = append(operations, yamlpatch.Operation{
				Type: yamlpatch.OperationAdd,
				Path: yamlpatch.MustParsePath(envPath),
				Value: map[string]interface{}{
					"env": map[string]interface{}{
						envVarName: fmt.Sprintf("${{ %s }}", cleanExpr),
					},
				},
			})

			modified, err := yamlpatch.Apply([]byte(content), yamlpatch.Patch(operations))
			if err != nil {
				return "", false, err
			}

			return string(modified), true, nil
		},
	}
}

func ADES101Fix(
	expression string,
	jobID string,
	stepIndex int,
	script string,
) Fix {

	expressionCopy := strings.TrimSpace(expression)
	scriptCopy := script

	return Fix{
		Title: "Move template expression to environment variable",
		Description: fmt.Sprintf(
			"Move template expression (%s) from with.script to an environment variable "+
				"to prevent template injection vulnerabilities.",
			expression,
		),

		Apply: func(content string) (string, bool, error) {
			var operations []yamlpatch.Operation

			var cleanExpr, fullExpr string
			if strings.HasPrefix(expressionCopy, "${{") && strings.HasSuffix(expressionCopy, "}}") {
				cleanExpr = strings.TrimSpace(expressionCopy[3 : len(expressionCopy)-2])
				fullExpr = expressionCopy
			} else {
				cleanExpr = expressionCopy
				fullExpr = fmt.Sprintf("${{ %s }}", cleanExpr)
			}

			envVarName := GenerateProcessedEnvVarName(cleanExpr)

			if !strings.Contains(scriptCopy, fullExpr) {
				return content, false, nil
			}

			newScript := strings.ReplaceAll(
				scriptCopy,
				fullExpr,
				fmt.Sprintf("$%s", envVarName),
			)

			scriptPath := fmt.Sprintf(
				"/jobs/%s/steps/%d/with/script",
				jobID,
				stepIndex,
			)
			operations = append(operations, yamlpatch.Operation{
				Type:  yamlpatch.OperationReplace,
				Path:  yamlpatch.MustParsePath(scriptPath),
				Value: newScript,
			})

			stepPath := fmt.Sprintf("/jobs/%s/steps/%d", jobID, stepIndex)
			operations = append(operations, yamlpatch.Operation{
				Type: yamlpatch.OperationAdd,
				Path: yamlpatch.MustParsePath(stepPath),
				Value: map[string]interface{}{
					"env": map[string]interface{}{
						envVarName: fmt.Sprintf("${{ %s }}", cleanExpr),
					},
				},
			})

			modified, err := yamlpatch.Apply([]byte(content), yamlpatch.Patch(operations))
			if err != nil {
				return "", false, err
			}

			return string(modified), true, nil
		},
	}
}

func ADES102Fix(
	expression string,
	jobID string,
	stepIndex int,
	script string,
) Fix {

	expressionCopy := strings.TrimSpace(expression)
	scriptCopy := script

	return Fix{
		Title: "Move template expression to environment variable",
		Description: fmt.Sprintf(
			"Move template expression (%s) from issue-close-message to an environment "+
				"variable to prevent template injection vulnerabilities.",
			expression,
		),

		Apply: func(content string) (string, bool, error) {
			var operations []yamlpatch.Operation

			var cleanExpr, fullExpr string
			if strings.HasPrefix(expressionCopy, "${{") && strings.HasSuffix(expressionCopy, "}}") {
				cleanExpr = strings.TrimSpace(expressionCopy[3 : len(expressionCopy)-2])
				fullExpr = expressionCopy
			} else {
				cleanExpr = expressionCopy
				fullExpr = fmt.Sprintf("${{ %s }}", cleanExpr)
			}

			envVarName := GenerateProcessedEnvVarName(cleanExpr)

			if !strings.Contains(scriptCopy, fullExpr) {
				return content, false, nil
			}

			newScript := strings.ReplaceAll(
				scriptCopy,
				fullExpr,
				fmt.Sprintf("$%s", envVarName),
			)

			withPath := fmt.Sprintf(
				"/jobs/%s/steps/%d/with/issue-close-message",
				jobID,
				stepIndex,
			)
			operations = append(operations, yamlpatch.Operation{
				Type:  yamlpatch.OperationReplace,
				Path:  yamlpatch.MustParsePath(withPath),
				Value: newScript,
			})

			stepPath := fmt.Sprintf("/jobs/%s/steps/%d", jobID, stepIndex)
			operations = append(operations, yamlpatch.Operation{
				Type: yamlpatch.OperationAdd,
				Path: yamlpatch.MustParsePath(stepPath),
				Value: map[string]interface{}{
					"env": map[string]interface{}{
						envVarName: fmt.Sprintf("${{ %s }}", cleanExpr),
					},
				},
			})

			modified, err := yamlpatch.Apply([]byte(content), yamlpatch.Patch(operations))
			if err != nil {
				return "", false, err
			}

			return string(modified), true, nil
		},
	}
}

func ADES103Fix(
	expression string,
	jobID string,
	stepIndex int,
	script string,
) Fix {

	expressionCopy := strings.TrimSpace(expression)
	scriptCopy := script

	return Fix{
		Title: "Move template expression to environment variable",
		Description: fmt.Sprintf(
			"Move template expression (%s) from pr-close-message to an environment "+
				"variable to prevent template injection vulnerabilities.",
			expression,
		),

		Apply: func(content string) (string, bool, error) {
			var operations []yamlpatch.Operation

			var cleanExpr, fullExpr string
			if strings.HasPrefix(expressionCopy, "${{") && strings.HasSuffix(expressionCopy, "}}") {
				cleanExpr = strings.TrimSpace(expressionCopy[3 : len(expressionCopy)-2])
				fullExpr = expressionCopy
			} else {
				cleanExpr = expressionCopy
				fullExpr = fmt.Sprintf("${{ %s }}", cleanExpr)
			}

			envVarName := GenerateProcessedEnvVarName(cleanExpr)

			if !strings.Contains(scriptCopy, fullExpr) {
				return content, false, nil
			}

			newScript := strings.ReplaceAll(
				scriptCopy,
				fullExpr,
				fmt.Sprintf("$%s", envVarName),
			)

			withPath := fmt.Sprintf(
				"/jobs/%s/steps/%d/with/pr-close-message",
				jobID,
				stepIndex,
			)
			operations = append(operations, yamlpatch.Operation{
				Type:  yamlpatch.OperationReplace,
				Path:  yamlpatch.MustParsePath(withPath),
				Value: newScript,
			})

			stepPath := fmt.Sprintf("/jobs/%s/steps/%d", jobID, stepIndex)
			operations = append(operations, yamlpatch.Operation{
				Type: yamlpatch.OperationAdd,
				Path: yamlpatch.MustParsePath(stepPath),
				Value: map[string]interface{}{
					"env": map[string]interface{}{
						envVarName: fmt.Sprintf("${{ %s }}", cleanExpr),
					},
				},
			})

			modified, err := yamlpatch.Apply([]byte(content), yamlpatch.Patch(operations))
			if err != nil {
				return "", false, err
			}

			return string(modified), true, nil
		},
	}
}

func ADES104Fix(
	expression string,
	jobID string,
	stepIndex int,
	script string,
) Fix {
	expressionCopy := strings.TrimSpace(expression)
	scriptCopy := script
	return Fix{
		Title: "Move template expression to environment variable",
		Description: fmt.Sprintf(
			"Move template expression (%s) from issue-title to an environment "+
				"variable to prevent template injection vulnerabilities.",
			expression,
		),
		Apply: func(content string) (string, bool, error) {
			var operations []yamlpatch.Operation
			var cleanExpr, fullExpr string
			if strings.HasPrefix(expressionCopy, "${{") && strings.HasSuffix(expressionCopy, "}}") {
				cleanExpr = strings.TrimSpace(expressionCopy[3 : len(expressionCopy)-2])
				fullExpr = expressionCopy
			} else {
				cleanExpr = expressionCopy
				fullExpr = fmt.Sprintf("${{ %s }}", cleanExpr)
			}

			envVarName := GenerateEnvVarName(cleanExpr)
			if !strings.Contains(scriptCopy, fullExpr) {
				return content, false, nil
			}
			newScript := strings.ReplaceAll(
				scriptCopy,
				fullExpr,
				fmt.Sprintf("$%s", envVarName),
			)
			withPath := fmt.Sprintf(
				"/jobs/%s/steps/%d/with/cmd",
				jobID,
				stepIndex,
			)
			operations = append(operations, yamlpatch.Operation{
				Type:  yamlpatch.OperationReplace,
				Path:  yamlpatch.MustParsePath(withPath),
				Value: newScript,
			})
			stepPath := fmt.Sprintf("/jobs/%s/steps/%d", jobID, stepIndex)
			operations = append(operations, yamlpatch.Operation{
				Type: yamlpatch.OperationAdd,
				Path: yamlpatch.MustParsePath(stepPath),
				Value: map[string]interface{}{
					"env": map[string]interface{}{
						envVarName: fmt.Sprintf("${{ %s }}", cleanExpr),
					},
				},
			})
			modified, err := yamlpatch.Apply([]byte(content), yamlpatch.Patch(operations))
			if err != nil {
				return "", false, err
			}
			return string(modified), true, nil
		},
	}
}

func ADES105Fix(
	expression string,
	jobID string,
	stepIndex int,
	runInput string,
) Fix {

	expressionCopy := strings.TrimSpace(expression)
	runCopy := runInput

	return Fix{
		Title: "Remove unsafe template expression from docker run input",
		Description: fmt.Sprintf(
			"Remove template expression (%s) from the run input of "+
				"addnab/docker-run-action. There is no safe way to use untrusted "+
				"expressions in this context without risking command injection.",
			expression,
		),

		Apply: func(content string) (string, bool, error) {
			var operations []yamlpatch.Operation

			// Normalize expression
			var fullExpr string
			if strings.HasPrefix(expressionCopy, "${{") && strings.HasSuffix(expressionCopy, "}}") {
				fullExpr = expressionCopy
			} else {
				fullExpr = fmt.Sprintf("${{ %s }}", expressionCopy)
			}

			// If the expression is not present, do nothing
			if !strings.Contains(runCopy, fullExpr) {
				return content, false, nil
			}

			// Remove the expression entirely
			newRun := strings.ReplaceAll(runCopy, fullExpr, "")

			// Replace the run input
			runPath := fmt.Sprintf(
				"/jobs/%s/steps/%d/with/run",
				jobID,
				stepIndex,
			)
			operations = append(operations, yamlpatch.Operation{
				Type:  yamlpatch.OperationReplace,
				Path:  yamlpatch.MustParsePath(runPath),
				Value: strings.TrimSpace(newRun),
			})

			// Apply patch
			modified, err := yamlpatch.Apply(
				[]byte(content),
				yamlpatch.Patch(operations),
			)
			if err != nil {
				return "", false, err
			}

			return string(modified), true, nil
		},
	}
}
