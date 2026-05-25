package domain

import "strings"

// Optional controlled edition value.
type Edition struct {
	value     string
	specified bool
}

// Creates an edition, treating blank input as unspecified.
func NewEdition(value string) (Edition, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Edition{}, nil
	}
	return Edition{value: value, specified: true}, nil
}

// Reports whether the edition has a non-blank value.
func (edition Edition) IsSpecified() bool {
	return edition.specified
}

// Returns the normalized edition, or an empty string when unspecified.
func (edition Edition) String() string {
	return edition.value
}
