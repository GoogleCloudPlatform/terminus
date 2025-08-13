package layout

import (
	"regexp"
	"unicode/utf8"
)

// ansiRegex matches ANSI escape sequences
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// visibleLength returns the visible length of a string (excluding ANSI escape sequences)
func visibleLength(s string) int {
	// Strip ANSI escape sequences
	cleaned := ansiRegex.ReplaceAllString(s, "")
	// Count runes, not bytes
	return utf8.RuneCountInString(cleaned)
}

// stripANSI removes all ANSI escape sequences from a string
func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}