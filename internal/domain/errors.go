package domain

import "fmt"

// Domain validation failure for a specific field.
type ValidationError struct {
	Field  string
	Reason string
}

// Returns a compact field-qualified validation message.
func (err ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", err.Field, err.Reason)
}

func validationError(field, reason string) error {
	return ValidationError{Field: field, Reason: reason}
}
