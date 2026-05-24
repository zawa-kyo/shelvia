package domain

import "strings"

// Non-empty book title.
type Title struct {
	value string
}

// Creates a title after trimming surrounding whitespace.
func NewTitle(value string) (Title, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Title{}, validationError("title", "required")
	}
	return Title{value: value}, nil
}

// Returns the normalized title text.
func (title Title) String() string {
	return title.value
}
