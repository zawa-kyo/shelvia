package domain

import "strings"

// Text field where blank input is treated as absent.
type OptionalText struct {
	value     string
	specified bool
}

// Creates optional text after trimming surrounding whitespace.
func NewOptionalText(value string) OptionalText {
	value = strings.TrimSpace(value)
	if value == "" {
		return OptionalText{}
	}
	return OptionalText{value: value, specified: true}
}

// Reports whether the optional text has a non-blank value.
func (text OptionalText) IsSpecified() bool {
	return text.specified
}

// Returns the normalized text, or an empty string when unspecified.
func (text OptionalText) String() string {
	return text.value
}
