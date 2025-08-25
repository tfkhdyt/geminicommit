package stringtools

import "strings"

// Subslice extracts a substring from a string that starts with the first occurrence of 'a'
// and ends with the last occurrence of 'b'. It removes all content before 'a' and after 'b'.
//
// Parameters:
// - value: The string to search within.
// - a: The starting delimiter.
// - b: The ending delimiter.
//
// Returns:
// The substring containing 'a', all content between 'a' and 'b', and 'b'.
// If either delimiter is not found, or if 'a' is not found before 'b', an empty string is returned.
func Subslice(value string, a string, b string) string {
	// Find the index of the first occurrence of 'a'.
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}

	// Find the index of the last occurrence of 'b'.
	posLast := strings.LastIndex(value, b)
	if posLast == -1 {
		return ""
	}

	// Ensure the start delimiter is found before the end delimiter.
	if posFirst >= posLast {
		return ""
	}

	// Extract the substring including both delimiters.
	// The slice starts at posFirst and ends at posLast + length of 'b'.
	return value[posFirst : posLast+len(b)]
}
