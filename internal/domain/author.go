package domain

import "strings"

// Non-empty book author name.
type Author struct {
	value string
}

// Creates an author after trimming surrounding whitespace.
func NewAuthor(value string) (Author, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Author{}, validationError("author", "required")
	}
	return Author{value: value}, nil
}

// Returns the normalized author name.
func (author Author) String() string {
	return author.value
}
