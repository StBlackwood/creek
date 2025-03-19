package utils

import "strings"

// TrimInput sanitizes user input by trimming spaces
func TrimInput(input string) string {
	return strings.TrimSpace(input)
}
