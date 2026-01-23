package ui

import "strings"

const (
	separatorChar  = "═"
	separatorCount = 35
)

// Separator returns a repeated separator line.
func Separator() string {
	return strings.Repeat(separatorChar, separatorCount)
}
