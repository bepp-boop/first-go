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

// ReplacementStyle defines how the expression should be replaced
type ReplacementStyle int

const (
	ADES100 ReplacementStyle = iota
	ADES101
	ADES102
	ADES103
	ADES104
	ADES105
	ADES106
	ADES107
	ADES108
	ADES109
	ADES110
	ADES111
	ADES112
	ADES113
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

func GenerateProcessedEnvVarName(expr string) string {
	return "process.env." + GenerateEnvVarName(expr)
}

func GenerateDotEnvVarName(expr string) string {
	return "env." + GenerateEnvVarName(expr)
}

func GenerateOsEnvVarName(expr string) string {
	return "os.getenv." + GenerateEnvVarName(expr)
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

// Shared helper function
func applyExpressionFix(
	content, expression, jobID string, stepIndex int,
	inputKey string, scriptCopy string, style ReplacementStyle,
) (string, bool, error) {

	expr := strings.TrimSpace(expression)
	cleanExpr := expr
	fullExpr := expr

	if strings.HasPrefix(expr, "${{") && strings.HasSuffix(expr, "}}") {
		cleanExpr = strings.TrimSpace(expr[3 : len(expr)-2])
		fullExpr = expr
	} else {
		fullExpr = fmt.Sprintf("${{ %s }}", cleanExpr)
	}

	envVarName := GenerateEnvVarName(cleanExpr)

	if !strings.Contains(scriptCopy, fullExpr) {
		return content, false, nil
	}

	// Determine replacement text
	var replacement string
	switch style {
	case ADES100, ADES104, ADES108, ADES111, ADES112:
		replacement = fmt.Sprintf("$%s", envVarName)
	case ADES109:
		replacement = fmt.Sprintf("{os.getenv('%s')}", envVarName)
	case ADES106:
		replacement = fmt.Sprintf("env.%s", envVarName)
	case ADES101, ADES102, ADES103:
		replacement = fmt.Sprintf("${process.env.%s}", envVarName)
	case ADES107:
		replacement = fmt.Sprintf("process.env.%s", envVarName)
	case ADES110, ADES113:
		replacement = fmt.Sprintf("$env:%s", envVarName)
	}

	newScript := scriptCopy

	switch style {
	default:
		newScript = strings.ReplaceAll(newScript, fullExpr, replacement)
	case ADES100:
		newScript = strings.ReplaceAll(newScript, fullExpr, replacement)
	case ADES101:
		// Match any single-quoted string that contains the template expression
		re := regexp.MustCompile(`'([^']*\$\{\{\s*` + regexp.QuoteMeta(cleanExpr) + `\s*\}\}[^']*)'`)
		newScript = re.ReplaceAllStringFunc(newScript, func(s string) string {
			// Remove surrounding single quotes
			inner := s[1 : len(s)-1]
			// Replace the expression with processed env var
			inner = strings.ReplaceAll(inner, fullExpr, replacement)
			fmt.Println(inner)
			// Wrap with backticks
			return "`" + inner + "`"
		})
	case ADES104:
		// Add double quotes around the full expression if not already present
		re := regexp.MustCompile(`(?m)([^"]|^)(` + regexp.QuoteMeta(fullExpr) + `)([^"]|$)`)
		newScript = re.ReplaceAllStringFunc(newScript, func(s string) string {
			return strings.ReplaceAll(s, fullExpr, fmt.Sprintf(`"%s"`, replacement))
		})
	case ADES107:
		// Remove quotation marks around the full expression if present
		re := regexp.MustCompile(`["']` + regexp.QuoteMeta(fullExpr) + `["']`)
		newScript = re.ReplaceAllString(newScript, replacement)

	}

	var operations []yamlpatch.Operation

	var inputPath string

	if inputKey == "run" {
		inputPath = fmt.Sprintf(
			"/jobs/%s/steps/%d/run",
			jobID,
			stepIndex,
		)
	} else {
		inputPath = fmt.Sprintf(
			"/jobs/%s/steps/%d/with/%s",
			jobID,
			stepIndex,
			inputKey,
		)
	}

	operations = append(operations, yamlpatch.Operation{
		Type:  yamlpatch.OperationReplace,
		Path:  yamlpatch.MustParsePath(inputPath),
		Value: newScript,
	})

	// Add env var to step
	stepPath := fmt.Sprintf("/jobs/%s/steps/%d/env", jobID, stepIndex)
	operations = append(operations, yamlpatch.Operation{
		Type: yamlpatch.OperationAdd,
		Path: yamlpatch.MustParsePath(stepPath),
		Value: map[string]interface{}{
			envVarName: fmt.Sprintf("${{ %s }}", cleanExpr),
		},
	})

	modified, err := yamlpatch.Apply([]byte(content), yamlpatch.Patch(operations))
	if err != nil {
		return "", false, err
	}
	return string(modified), true, nil
}

// ----------------------- Refactored ADESxxxFix Functions -----------------------
func ADES100Fix(expression, jobID string, stepIndex int, script string) Fix {
	return Fix{
		Title:       "Fix template expression in run: directive",
		Description: fmt.Sprintf("Move template expression (%s) from run to an environment variable", expression),
		Apply: func(content string) (string, bool, error) {
			return applyExpressionFix(content, expression, jobID, stepIndex, "run", script, ADES100)
		},
	}
}

func ADES101Fix(expression, jobID string, stepIndex int, script string) Fix {
	return Fix{
		Title:       "Fix template expression in action/github-script action",
		Description: fmt.Sprintf("Move template expression (%s) from with.script to an environment variable", expression),
		Apply: func(content string) (string, bool, error) {
			return applyExpressionFix(content, expression, jobID, stepIndex, "script", script, ADES101)
		},
	}
}

func ADES102Fix(expression, jobID string, stepIndex int, script string) Fix {
	return Fix{
		Title:       "Fix template expression in issue-close-message input of root/issues-closer-action",
		Description: fmt.Sprintf("Move template expression (%s) from issue-close-message to an environment variable", expression),
		Apply: func(content string) (string, bool, error) {
			return applyExpressionFix(content, expression, jobID, stepIndex, "issue-close-message", script, ADES102)
		},
	}
}

func ADES103Fix(expression, jobID string, stepIndex int, script string) Fix {
	return Fix{
		Title:       "Fix template expression in pr-close-message input of root/issues-closer-action",
		Description: fmt.Sprintf("Move template expression (%s) from pr-close-message to an environment variable", expression),
		Apply: func(content string) (string, bool, error) {
			return applyExpressionFix(content, expression, jobID, stepIndex, "pr-close-message", script, ADES103)
		},
	}
}

func ADES104Fix(expression, jobID string, stepIndex int, script string) Fix {
	return Fix{
		Title:       "Fix template expression in cmd input in sergeysova/jq-action",
		Description: fmt.Sprintf("Move template expression (%s) from cmd to an environment variable", expression),
		Apply: func(content string) (string, bool, error) {
			return applyExpressionFix(content, expression, jobID, stepIndex, "cmd", script, ADES104)
		},
	}
}

func ADES105Fix(expression, jobID string, stepIndex int, runInput string) Fix {
	return Fix{
		Title:       "Remove unsafe template expression from docker run input",
		Description: fmt.Sprintf("Remove template expression (%s) from run input of addnab/docker-run-action", expression),
		Apply: func(content string) (string, bool, error) {
			expr := strings.TrimSpace(expression)
			fullExpr := expr
			if !strings.HasPrefix(expr, "${{") || !strings.HasSuffix(expr, "}}") {
				fullExpr = fmt.Sprintf("${{ %s }}", expr)
			}
			if !strings.Contains(runInput, fullExpr) {
				return content, false, nil
			}
			newRun := strings.ReplaceAll(runInput, fullExpr, "")
			runPath := fmt.Sprintf("/jobs/%s/steps/%d/with/run", jobID, stepIndex)
			operations := []yamlpatch.Operation{{
				Type:  yamlpatch.OperationReplace,
				Path:  yamlpatch.MustParsePath(runPath),
				Value: strings.TrimSpace(newRun),
			}}
			modified, err := yamlpatch.Apply([]byte(content), yamlpatch.Patch(operations))
			if err != nil {
				return "", false, err
			}
			return string(modified), true, nil
		},
	}
}

func ADES106Fix(expression, jobID string, stepIndex int, script string) Fix {
	return Fix{
		Title:       "Fix template expression in expression input in cardinalby/js-eval-action",
		Description: fmt.Sprintf("Move template expression (%s) from with/expression to env", expression),
		Apply: func(content string) (string, bool, error) {
			return applyExpressionFix(content, expression, jobID, stepIndex, "expression", script, ADES106)
		},
	}
}

func ADES107Fix(expression, jobID string, stepIndex int, script string) Fix {
	return Fix{
		Title:       "Fix template expression in custom_payload input in 8398a7/action-slack",
		Description: fmt.Sprintf("Move template expression (%s) from custom_payload to env", expression),
		Apply: func(content string) (string, bool, error) {
			return applyExpressionFix(content, expression, jobID, stepIndex, "custom_payload", script, ADES107)
		},
	}
}

func ADES108Fix(expression, jobID string, stepIndex int, script string) Fix {
	return Fix{
		Title:       "Fix template expression in script input of appleboy/ssh-action",
		Description: fmt.Sprintf("Move template expression (%s) from script to env", expression),
		Apply: func(content string) (string, bool, error) {
			return applyExpressionFix(content, expression, jobID, stepIndex, "script", script, ADES108)
		},
	}
}

func ADES109Fix(expression, jobID string, stepIndex int, script string) Fix {
	return Fix{
		Title:       "Fix template expression in script input of jannekem/run-python-script-action",
		Description: fmt.Sprintf("Move template expression (%s) from script to env", expression),
		Apply: func(content string) (string, bool, error) {
			return applyExpressionFix(content, expression, jobID, stepIndex, "script", script, ADES109)
		},
	}
}

func ADES110Fix(expression, jobID string, stepIndex int, script string) Fix {
	return Fix{
		Title:       "Fix template expression in script input of Amadevus/pwsh-script",
		Description: fmt.Sprintf("Move template expression (%s) from script to env", expression),
		Apply: func(content string) (string, bool, error) {
			return applyExpressionFix(content, expression, jobID, stepIndex, "script", script, ADES110)
		},
	}
}

func ADES111Fix(expression, jobID string, stepIndex int, script string) Fix {
	return Fix{
		Title:       "Fix template expression in cmd input of mikefarah/yg",
		Description: fmt.Sprintf("Move template expression (%s) from cmd to env", expression),
		Apply: func(content string) (string, bool, error) {
			return applyExpressionFix(content, expression, jobID, stepIndex, "cmd", script, ADES111)
		},
	}
}

func ADES112Fix(expression, jobID string, stepIndex int, script string) Fix {
	return Fix{
		Title:       "Fix template expression in cmd input of devorbitus/yg-action-output",
		Description: fmt.Sprintf("Move template expression (%s) from cmd to env", expression),
		Apply: func(content string) (string, bool, error) {
			return applyExpressionFix(content, expression, jobID, stepIndex, "cmd", script, ADES112)
		},
	}
}

func ADES113Fix(expression, jobID string, stepIndex int, script string) Fix {
	return Fix{
		Title:       "Fix template expression in expression in inLineScript of azure/powershell",
		Description: fmt.Sprintf("Move template expression (%s) from inlineScript to env", expression),
		Apply: func(content string) (string, bool, error) {
			return applyExpressionFix(content, expression, jobID, stepIndex, "inlineScript", script, ADES113)
		},
	}
}
