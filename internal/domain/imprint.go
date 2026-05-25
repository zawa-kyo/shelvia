package domain

import "strings"

// Optional controlled imprint value.
type Imprint struct {
	value     string
	specified bool
}

// Creates an imprint, treating blank input as unspecified.
func NewImprint(value string) (Imprint, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Imprint{}, nil
	}
	return Imprint{value: value, specified: true}, nil
}

// Reports whether the imprint has a non-blank value.
func (imprint Imprint) IsSpecified() bool {
	return imprint.specified
}

// Returns the normalized imprint, or an empty string when unspecified.
func (imprint Imprint) String() string {
	return imprint.value
}
