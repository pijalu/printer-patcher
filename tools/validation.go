package tools

import (
	"regexp"
	"strings"
)

// ValidateOutput checks if the output matches the expected pattern
func ValidateOutput(output, expected string) bool {
	if expected == "" {
		// If no expected output is defined, consider it successful
		return true
	}

	// Trim whitespace from both output and expected
	output = strings.TrimSpace(output)
	expected = strings.TrimSpace(expected)

	// Try to match as a regular expression
	matched, err := regexp.MatchString(expected, output)
	if err == nil {
		return matched
	}

	// If regex fails, do a simple string comparison
	return output == expected
}
