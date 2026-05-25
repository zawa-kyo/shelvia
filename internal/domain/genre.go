package domain

import "strings"

// Non-empty controlled genre value.
type Genre struct {
	value string
}

// Creates a genre after trimming surrounding whitespace.
func NewGenre(value string) (Genre, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Genre{}, validationError("genre", "required")
	}
	return Genre{value: value}, nil
}

// Returns the normalized genre.
func (genre Genre) String() string {
	return genre.value
}
