package domain

import "strings"

// Source file path used for diagnostics.
type FilePath struct {
	value string
}

// Creates a file path after trimming surrounding whitespace.
func NewFilePath(value string) FilePath {
	return FilePath{value: strings.TrimSpace(value)}
}

// Returns the normalized file path.
func (path FilePath) String() string {
	return path.value
}
