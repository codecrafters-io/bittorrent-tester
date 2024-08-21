package internal

import (
	"strings"
)

// Logger treats % characters as formatting parameters
// Escape them by using double %
func escape(input string) string {
	return strings.ReplaceAll(input, "%", "%%")
}