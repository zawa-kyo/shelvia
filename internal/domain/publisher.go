package domain

import "strings"

// Non-empty controlled publisher value.
type Publisher struct {
	value string
}

// Creates a publisher after trimming surrounding whitespace.
func NewPublisher(value string) (Publisher, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Publisher{}, validationError("publisher", "required")
	}
	return Publisher{value: value}, nil
}

// Returns the normalized publisher name.
func (publisher Publisher) String() string {
	return publisher.value
}
