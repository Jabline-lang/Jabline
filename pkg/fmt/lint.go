package fmt

import (
	"fmt"
	"strings"
	"unicode"
)

// LintIssue represents a single linting issue found in source code.
type LintIssue struct {
	Line    int
	Column  int
	Message string
	Severity string // "warning" or "error"
	Rule    string
}

// LintResult contains all linting issues for a file.
type LintResult struct {
	Filename string
	Issues   []LintIssue
}

// Lint checks a source file for style and correctness issues.
// Returns a list of lint issues. An empty list means no issues found.
func Lint(filename, source string) []LintIssue {
	var issues []LintIssue
	lines := strings.Split(source, "\n")

	for i, line := range lines {
		lineNum := i + 1

		// R1: Tab indentation (tabs required, spaces forbidden for indentation)
		leading := leadingWhitespace(line)
		if strings.Contains(leading, "  ") && !strings.Contains(leading, "\t") {
			issues = append(issues, LintIssue{
				Line:     lineNum,
				Column:   1,
				Message:  "use tabs for indentation, not spaces",
				Severity: "warning",
				Rule:     "indent-tabs",
			})
		}

		// Skip empty lines and comments for remaining checks
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// R2: Trailing whitespace
		if len(line) > 0 && (line[len(line)-1] == ' ' || line[len(line)-1] == '\t') {
			col := len(line) - len(strings.TrimRight(line, " \t")) + 1
			issues = append(issues, LintIssue{
				Line:     lineNum,
				Column:   col,
				Message:  "trailing whitespace",
				Severity: "warning",
				Rule:     "no-trailing-spaces",
			})
		}

		// R3: Line length > 120
		if len(line) > 120 {
			issues = append(issues, LintIssue{
				Line:     lineNum,
				Column:   121,
				Message:  fmt.Sprintf("line too long (%d > 120 characters)", len(line)),
				Severity: "warning",
				Rule:     "max-line-length",
			})
		}

		// R4: Naming conventions (camelCase for variables/functions)
		if strings.Contains(trimmed, "let ") || strings.Contains(trimmed, "const ") {
			for _, part := range strings.Fields(trimmed) {
				if strings.HasPrefix(part, "let ") || strings.HasPrefix(part, "const ") {
					continue
				}
				// Check for snake_case in identifiers
				if strings.Contains(part, "_") && part == strings.ToLower(part) {
					name := strings.TrimRight(part, ";,=:({[]})")
					if isIdentifier(name) && !isAllowedSnakeCase(name) {
						issues = append(issues, LintIssue{
							Line:     lineNum,
							Column:   strings.Index(line, name) + 1,
							Message:  fmt.Sprintf("'%s' should use camelCase naming", name),
							Severity: "warning",
							Rule:     "camelcase",
						})
					}
				}
			}
		}

		// R5: Semicolons at end of line (unnecessary)
		trimmed = strings.TrimSpace(line)
		if strings.HasSuffix(trimmed, ";") {
			codePart := trimmed[:len(trimmed)-1]
			if !strings.Contains(codePart, "\"") && !strings.Contains(codePart, "'") {
				issues = append(issues, LintIssue{
					Line:     lineNum,
					Column:   len(trimmed),
					Message:  "unnecessary semicolon",
					Severity: "warning",
					Rule:     "no-semicolons",
				})
			}
		}
	}

	return issues
}

// FormatLintResult formats lint results for human-readable output.
func FormatLintResult(result LintResult) string {
	var sb strings.Builder
	for _, issue := range result.Issues {
		severity := "W"
		if issue.Severity == "error" {
			severity = "E"
		}
		sb.WriteString(fmt.Sprintf("%s:%d:%d: [%s] %s (%s)\n",
			result.Filename, issue.Line, issue.Column, severity, issue.Message, issue.Rule))
	}
	return sb.String()
}

func leadingWhitespace(line string) string {
	for i, r := range line {
		if r != ' ' && r != '\t' {
			return line[:i]
		}
	}
	return line
}

func isIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, r := range s {
		if i == 0 && !unicode.IsLetter(r) && r != '_' {
			return false
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}

func isAllowedSnakeCase(name string) bool {
	// Allow single underscores, leading underscores (private), and common patterns
	if name == "_" || strings.HasPrefix(name, "__") {
		return true
	}
	// Allow uppercase constants
	if name == strings.ToUpper(name) && strings.Contains(name, "_") {
		return true
	}
	return false
}
