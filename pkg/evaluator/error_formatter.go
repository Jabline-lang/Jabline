package evaluator

import (
	"fmt"
	"strings"

	"jabline/pkg/token"
)

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
)

// ErrorLevel represents the severity of an error
type ErrorLevel int

const (
	ErrorLevelError ErrorLevel = iota
	ErrorLevelWarning
	ErrorLevelInfo
)

// FormattedError represents an error with formatting information
type FormattedError struct {
	Level      ErrorLevel
	Message    string
	Line       int
	Column     int
	Filename   string
	SourceLine string
	Suggestion string
}

// ErrorFormatter handles the formatting and display of errors
type ErrorFormatter struct {
	UseColors bool
	ShowCode  bool
}

// NewErrorFormatter creates a new error formatter
func NewErrorFormatter(useColors, showCode bool) *ErrorFormatter {
	return &ErrorFormatter{
		UseColors: useColors,
		ShowCode:  showCode,
	}
}

// FormatError formats a single error with colors and context
func (ef *ErrorFormatter) FormatError(err FormattedError) string {
	var result strings.Builder

	// Error level indicator
	levelIndicator := ef.getLevelIndicator(err.Level)
	levelColor := ef.getLevelColor(err.Level)

	if ef.UseColors {
		result.WriteString(levelColor + ColorBold)
	}

	// Header line
	result.WriteString(fmt.Sprintf("%s ", levelIndicator))

	if err.Filename != "" {
		result.WriteString(fmt.Sprintf("%s:", err.Filename))
	}

	if err.Line > 0 {
		result.WriteString(fmt.Sprintf("%d:", err.Line))
	}

	if err.Column > 0 {
		result.WriteString(fmt.Sprintf("%d:", err.Column))
	}

	result.WriteString(" " + err.Message)

	if ef.UseColors {
		result.WriteString(ColorReset)
	}
	result.WriteString("\n")

	// Show source code context if available
	if ef.ShowCode && err.SourceLine != "" {
		result.WriteString(ef.formatSourceLine(err))
	}

	// Show suggestion if available
	if err.Suggestion != "" {
		result.WriteString(ef.formatSuggestion(err.Suggestion))
	}

	return result.String()
}

// FormatMultipleErrors formats multiple errors
func (ef *ErrorFormatter) FormatMultipleErrors(errors []FormattedError, filename string) string {
	if len(errors) == 0 {
		return ""
	}

	var result strings.Builder

	// Summary header
	if ef.UseColors {
		result.WriteString(ColorRed + ColorBold)
	}

	errorCount := 0
	warningCount := 0

	for _, err := range errors {
		if err.Level == ErrorLevelError {
			errorCount++
		} else if err.Level == ErrorLevelWarning {
			warningCount++
		}
	}

	if errorCount > 0 && warningCount > 0 {
		result.WriteString(fmt.Sprintf("Found %d error(s) and %d warning(s) in %s\n",
			errorCount, warningCount, filename))
	} else if errorCount > 0 {
		result.WriteString(fmt.Sprintf("Found %d error(s) in %s\n", errorCount, filename))
	} else if warningCount > 0 {
		result.WriteString(fmt.Sprintf("Found %d warning(s) in %s\n", warningCount, filename))
	}

	if ef.UseColors {
		result.WriteString(ColorReset)
	}
	result.WriteString("\n")

	// Individual errors
	for i, err := range errors {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(ef.FormatError(err))
	}

	return result.String()
}

// formatSourceLine formats the source line with highlighting
func (ef *ErrorFormatter) formatSourceLine(err FormattedError) string {
	var result strings.Builder

	// Line number with padding
	lineNumStr := fmt.Sprintf("%d", err.Line)
	padding := strings.Repeat(" ", 4-len(lineNumStr))

	if ef.UseColors {
		result.WriteString(ColorBlue + ColorDim)
	}
	result.WriteString(fmt.Sprintf("%s%s | ", padding, lineNumStr))
	if ef.UseColors {
		result.WriteString(ColorReset)
	}

	// Source line
	result.WriteString(err.SourceLine)
	result.WriteString("\n")

	// Error pointer
	if err.Column > 0 {
		if ef.UseColors {
			result.WriteString(ColorBlue + ColorDim)
		}
		result.WriteString("     | ")
		if ef.UseColors {
			result.WriteString(ColorReset + ColorRed + ColorBold)
		}

		// Add spaces to point to the exact column
		spaces := strings.Repeat(" ", err.Column-1)
		result.WriteString(spaces + "^")

		if ef.UseColors {
			result.WriteString(ColorReset)
		}
		result.WriteString("\n")
	}

	return result.String()
}

// formatSuggestion formats a helpful suggestion
func (ef *ErrorFormatter) formatSuggestion(suggestion string) string {
	var result strings.Builder

	if ef.UseColors {
		result.WriteString(ColorGreen + ColorBold)
	}
	result.WriteString("💡 Suggestion: ")
	if ef.UseColors {
		result.WriteString(ColorReset + ColorGreen)
	}
	result.WriteString(suggestion)
	if ef.UseColors {
		result.WriteString(ColorReset)
	}
	result.WriteString("\n")

	return result.String()
}

// getLevelIndicator returns the indicator for error level
func (ef *ErrorFormatter) getLevelIndicator(level ErrorLevel) string {
	switch level {
	case ErrorLevelError:
		return "❌ ERROR"
	case ErrorLevelWarning:
		return "⚠️  WARNING"
	case ErrorLevelInfo:
		return "ℹ️  INFO"
	default:
		return "❓ UNKNOWN"
	}
}

// getLevelColor returns the color for error level
func (ef *ErrorFormatter) getLevelColor(level ErrorLevel) string {
	if !ef.UseColors {
		return ""
	}

	switch level {
	case ErrorLevelError:
		return ColorRed
	case ErrorLevelWarning:
		return ColorYellow
	case ErrorLevelInfo:
		return ColorBlue
	default:
		return ColorWhite
	}
}

// CreateParserError creates a formatted parser error
func CreateParserError(message string, tok token.Token, filename string, sourceLine string) FormattedError {
	suggestion := getSuggestionForParserError(message)

	return FormattedError{
		Level:      ErrorLevelError,
		Message:    message,
		Line:       tok.Line,
		Column:     tok.Column,
		Filename:   filename,
		SourceLine: sourceLine,
		Suggestion: suggestion,
	}
}

// CreateRuntimeError creates a formatted runtime error
func CreateRuntimeError(message string, line, column int, filename string, sourceLine string) FormattedError {
	suggestion := getSuggestionForRuntimeError(message)

	return FormattedError{
		Level:      ErrorLevelError,
		Message:    message,
		Line:       line,
		Column:     column,
		Filename:   filename,
		SourceLine: sourceLine,
		Suggestion: suggestion,
	}
}

// CreateWarning creates a formatted warning
func CreateWarning(message string, line, column int, filename string, sourceLine string) FormattedError {
	return FormattedError{
		Level:      ErrorLevelWarning,
		Message:    message,
		Line:       line,
		Column:     column,
		Filename:   filename,
		SourceLine: sourceLine,
	}
}

// getSuggestionForParserError provides helpful suggestions for parser errors
func getSuggestionForParserError(message string) string {
	if strings.Contains(message, "no prefix parse function") {
		if strings.Contains(message, "ILLEGAL") {
			return "Check for invalid characters or unsupported syntax"
		}
		if strings.Contains(message, ";") {
			return "Remove the semicolon or add an expression before it"
		}
		if strings.Contains(message, "}") {
			return "Check if you have unmatched opening braces"
		}
		if strings.Contains(message, ")") {
			return "Check if you have unmatched opening parentheses"
		}
		if strings.Contains(message, "]") {
			return "Check if you have unmatched opening brackets"
		}
	}

	if strings.Contains(message, "expected next token") {
		if strings.Contains(message, "RPAREN") {
			return "Add a closing parenthesis ')'"
		}
		if strings.Contains(message, "RBRACE") {
			return "Add a closing brace '}'"
		}
		if strings.Contains(message, "RBRACKET") {
			return "Add a closing bracket ']'"
		}
		if strings.Contains(message, "SEMICOLON") {
			return "Add a semicolon ';' at the end of the statement"
		}
	}

	if strings.Contains(message, "could not parse") && strings.Contains(message, "integer") {
		return "Check if the number format is correct (e.g., 123, not 123.45.67)"
	}

	if strings.Contains(message, "could not parse") && strings.Contains(message, "float") {
		return "Check if the decimal number format is correct (e.g., 123.45)"
	}

	return ""
}

// getSuggestionForRuntimeError provides helpful suggestions for runtime errors
func getSuggestionForRuntimeError(message string) string {
	if strings.Contains(message, "identifier not found") {
		return "Check if the variable is declared before using it"
	}

	if strings.Contains(message, "wrong number of arguments") {
		return "Check the function documentation for the correct number of parameters"
	}

	if strings.Contains(message, "division by zero") {
		return "Add a check to ensure the divisor is not zero"
	}

	if strings.Contains(message, "index out of range") {
		return "Check if the array index is within valid bounds"
	}

	if strings.Contains(message, "error parsing JSON") {
		return "Verify that the JSON string is properly formatted with quotes around keys and string values"
	}

	if strings.Contains(message, "cannot compute square root of negative number") {
		return "Ensure the number is positive before calculating square root"
	}

	if strings.Contains(message, "invalid regex pattern") {
		return "Check the regex syntax - you might need to escape special characters"
	}

	return ""
}

// ExtractSourceLine extracts the source line from input string
func ExtractSourceLine(input string, lineNumber int) string {
	if lineNumber <= 0 {
		return ""
	}

	lines := strings.Split(input, "\n")
	if lineNumber > len(lines) {
		return ""
	}

	return strings.TrimRight(lines[lineNumber-1], "\r\n")
}

// DetectCommonMistakes analyzes code for common mistakes and returns warnings
func DetectCommonMistakes(input string) []FormattedError {
	var warnings []FormattedError
	lines := strings.Split(input, "\n")

	for i, line := range lines {
		lineNum := i + 1
		trimmedLine := strings.TrimSpace(line)

		// Check for common mistakes
		if strings.Contains(line, "=") && strings.Count(line, "=") == 1 &&
			strings.Contains(line, "if") && !strings.Contains(line, "==") {
			warnings = append(warnings, CreateWarning(
				"Possible assignment in condition - did you mean '==' for comparison?",
				lineNum, strings.Index(line, "=")+1, "", line))
		}

		if strings.HasSuffix(trimmedLine, ";") &&
			(strings.HasPrefix(trimmedLine, "let ") ||
				strings.HasPrefix(trimmedLine, "const ") ||
				strings.HasPrefix(trimmedLine, "echo(")) {
			warnings = append(warnings, CreateWarning(
				"Unnecessary semicolon - Jabline doesn't require semicolons",
				lineNum, len(line), "", line))
		}

		if strings.Contains(line, "var ") {
			warnings = append(warnings, CreateWarning(
				"Use 'let' or 'const' instead of 'var' in Jabline",
				lineNum, strings.Index(line, "var ")+1, "", line))
		}
	}

	return warnings
}
