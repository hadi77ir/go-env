package env

import (
	"fmt"
	"os"
	"strings"
)

// ExpandEnv expands environment variables in the input string without using regex
// Supports the following formats:
// - $var
// - ${var}
// - ${var:-default}  (use default if var is unset or empty)
// - ${var:+alt}      (use alt if var is set and non-empty)
// - ${var:?error}    (error if var is unset or empty)
// - ${var:=default}  (set var to default if unset or empty, then use it)
func ExpandEnv(input string) (string, error) {
	var result strings.Builder
	i := 0

	for i < len(input) {
		if input[i] == '$' {
			// Found a potential variable
			expanded, newPos, err := parseVariable(input, i)
			if err != nil {
				return "", err
			}
			result.WriteString(expanded)
			i = newPos
		} else {
			// Regular character
			result.WriteByte(input[i])
			i++
		}
	}

	return result.String(), nil
}

// parseVariable parses a variable starting at position pos in the input string
// Returns the expanded value, the new position after the variable, and any error
func parseVariable(input string, pos int) (string, int, error) {
	if pos >= len(input) || input[pos] != '$' {
		return "", pos, fmt.Errorf("expected '$' at position %d", pos)
	}

	pos++ // Skip the '$'

	if pos >= len(input) {
		// Just a '$' at the end
		return "$", pos, nil
	}

	if input[pos] == '{' {
		// Handle ${...} format
		return parseBracedVariable(input, pos)
	} else {
		// Handle $var format
		return parseSimpleVariable(input, pos)
	}
}

// parseSimpleVariable parses a simple $var format
func parseSimpleVariable(input string, pos int) (string, int, error) {
	start := pos

	// Variable name must start with letter or underscore
	if pos >= len(input) || (!isLetter(input[pos]) && input[pos] != '_') {
		// Not a valid variable name, return the $ as literal
		return "$", pos, nil
	}

	// Continue while we have valid variable name characters (up to 64 chars max)
	for pos < len(input) && (isAlphaNum(input[pos]) || input[pos] == '_') && (pos-start) < 64 {
		pos++
	}

	varName := input[start:pos]
	if len(varName) == 0 || len(varName) > 64 {
		// Invalid variable name, return $ as literal
		return "$", start, nil
	}

	return os.Getenv(varName), pos, nil
}

// parseBracedVariable parses a ${...} format variable
func parseBracedVariable(input string, pos int) (string, int, error) {
	if pos >= len(input) || input[pos] != '{' {
		return "", pos, fmt.Errorf("expected '{' at position %d", pos)
	}

	pos++ // Skip the '{'
	start := pos

	// Find the closing brace
	braceCount := 1
	for pos < len(input) && braceCount > 0 {
		if input[pos] == '{' {
			braceCount++
		} else if input[pos] == '}' {
			braceCount--
		}
		if braceCount > 0 {
			pos++
		}
	}

	if braceCount > 0 {
		return "", pos, fmt.Errorf("unclosed brace in variable expression")
	}

	content := input[start:pos]
	pos++ // Skip the closing '}'

	expanded, err := expandBracedContent(content)
	if err != nil {
		return "", 0, err
	}
	return expanded, pos, nil
}

// expandBracedContent handles the expansion of content within braces
func expandBracedContent(content string) (string, error) {
	// Validate variable name in braced content
	var varName string

	// Look for parameter expansion operators
	if idx := strings.Index(content, ":-"); idx != -1 {
		// ${var:-default} - use default if var is unset or empty
		varName = content[:idx]
		if !isValidVarName(varName) {
			return fmt.Sprintf("${%s}", content), nil // Return as literal if invalid
		}
		defaultValue := content[idx+2:]
		if value := os.Getenv(varName); value != "" {
			return value, nil
		}
		return defaultValue, nil

	} else if idx := strings.Index(content, ":+"); idx != -1 {
		// ${var:+alt} - use alt if var is set and non-empty
		varName = content[:idx]
		if !isValidVarName(varName) {
			return fmt.Sprintf("${%s}", content), nil // Return as literal if invalid
		}
		altValue := content[idx+2:]
		if value := os.Getenv(varName); value != "" {
			return altValue, nil
		}
		return "", nil

	} else if idx := strings.Index(content, ":?"); idx != -1 {
		// ${var:?error} - error if var is unset or empty
		varName = content[:idx]
		if !isValidVarName(varName) {
			return fmt.Sprintf("${%s}", content), nil // Return as literal if invalid
		}
		errorMsg := content[idx+2:]
		if value := os.Getenv(varName); value != "" {
			return value, nil
		}
		return "", fmt.Errorf("variable '%s' is unset or empty: %s", varName, errorMsg)

	} else if idx := strings.Index(content, ":="); idx != -1 {
		// ${var:=default} - set var to default if unset or empty, then use it
		varName = content[:idx]
		if !isValidVarName(varName) {
			return fmt.Sprintf("${%s}", content), nil // Return as literal if invalid
		}
		defaultValue := content[idx+2:]
		if value := os.Getenv(varName); value != "" {
			return value, nil
		}
		// Set the environment variable to the default value
		os.Setenv(varName, defaultValue)
		return defaultValue, nil
	}

	// Simple ${var} format
	varName = content
	if !isValidVarName(varName) {
		return fmt.Sprintf("${%s}", content), nil // Return as literal if invalid
	}
	return os.Getenv(content), nil
}

// Helper functions for character classification
func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlphaNum(c byte) bool {
	return isLetter(c) || isDigit(c)
}

// isValidVarName validates environment variable name according to the rules:
// - Must be 1-64 characters long
// - Must start with a letter [A-Za-z] or underscore [_]
// - Can contain letters, digits, and underscores [A-Za-z0-9_]
func isValidVarName(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}

	// Must start with letter or underscore
	if !isLetter(name[0]) && name[0] != '_' {
		return false
	}

	// Rest of the characters must be alphanumeric or underscore
	for i := 1; i < len(name); i++ {
		if !isAlphaNum(name[i]) && name[i] != '_' {
			return false
		}
	}

	return true
}
