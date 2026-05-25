package domain

import "fmt"

// Integer score from 0 to 100.
type Rating struct {
	value int
}

// Creates a rating when the value is between 0 and 100 inclusive.
func NewRating(value int) (Rating, error) {
	if value < 0 || value > 100 {
		return Rating{}, validationError("rating", fmt.Sprintf("must be between 0 and 100: %d", value))
	}
	return Rating{value: value}, nil
}

// Returns the rating as an integer.
func (rating Rating) Int() int {
	return rating.value
}
