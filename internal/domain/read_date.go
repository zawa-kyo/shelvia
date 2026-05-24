package domain

import "time"

// Calendar date when a book was read.
type ReadDate struct {
	value time.Time
}

// Creates a read date and normalizes it to a UTC calendar date.
func NewReadDate(value time.Time) (ReadDate, error) {
	if value.IsZero() {
		return ReadDate{}, validationError("read_date", "required")
	}
	year, month, day := value.Date()
	return ReadDate{value: time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}, nil
}

// Returns the normalized UTC date value.
func (date ReadDate) Time() time.Time {
	return date.value
}

// Formats the read date as YYYY-MM-DD.
func (date ReadDate) String() string {
	return date.value.Format("2006-01-02")
}
