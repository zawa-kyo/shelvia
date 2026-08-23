package domain

import "time"

// Raw input used to construct a book.
type BookDraft struct {
	Title      string
	Author     string
	Rating     *int
	ReadDate   time.Time
	Genre      string
	Publisher  string
	Edition    string
	Imprint    string
	Series     string
	Translator string
	Summary    string
	Body       string
	FilePath   string
}
