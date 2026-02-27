package ui

import "strings"

const (
	separatorChar  = "-"
	separatorCount = 35
)

func Separator() string {
	return strings.Repeat(separatorChar, separatorCount)
}
