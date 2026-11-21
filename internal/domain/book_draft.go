package domain

import "time"

// Raw input used to construct a book.
type BookDraft struct {
	Authors     []string
	Genres      []string
	Translators []string
	Title       string

	Rating   *int
	ReadDate time.Time

	Publisher string
	Edition   string
	Imprint   string
	Series    string

	Summary  string
	Body     string
	FilePath string
}
