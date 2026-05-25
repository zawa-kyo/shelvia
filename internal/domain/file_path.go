package domain

import (
	"path"
	"strings"
)

// Source file path used for diagnostics.
type FilePath struct {
	value string
}

// Creates a slash-separated shelf path for diagnostics.
func NewFilePath(value string) FilePath {
	return FilePath{value: cleanShelfPath(value)}
}

// Reports whether a source path is known.
func (filePath FilePath) IsSpecified() bool {
	return filePath.value != ""
}

// Returns the last path element.
func (filePath FilePath) Base() string {
	if !filePath.IsSpecified() {
		return ""
	}
	return path.Base(filePath.value)
}

// Returns the normalized file path.
func (filePath FilePath) String() string {
	return filePath.value
}

func cleanShelfPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	cleaned := path.Clean(strings.ReplaceAll(trimmed, `\`, `/`))
	if cleaned == "." {
		return ""
	}
	return cleaned
}
