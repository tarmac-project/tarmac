/*
Package sanitize provides functions to sanitize user input into a safe format.
*/
package sanitize

import (
	"strings"
)

// String sanitizes a string by removing newline characters.
func String(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r", ""), "\n", "")
}
