package application

import (
	"time"

	"github.com/zawa-kyo/shelvia/internal/domain"
)

// Raw shelf data loaded from an external source.
type ShelfData struct {
	Config AllowedValuesData
	Books  []BookData
}

// Raw allowed values loaded from config.toml.
type AllowedValuesData struct {
	Genres     []string
	Publishers []string
	Imprints   []string
	Editions   []EditionData
}

// Raw edition setting loaded from config.toml.
type EditionData struct {
	Name            string
	ImprintRequired bool
}

// Raw book data loaded from a TOML book file.
type BookData struct {
	Path      string
	Title     string
	Author    string
	Rating    int
	ReadDate  time.Time
	Genre     string
	Publisher string
	Edition   string
	Imprint   string

	Series     string
	Translator string
	Summary    string
	Body       string
}

// Reads and writes shelf files without exposing filesystem details.
type ShelfRepository interface {
	Init(root string) error
	CreateBook(root, title string, readDate time.Time) (string, error)
	Load(root string) (ShelfData, error)
}

// Searches validated books without exposing the query backend.
type QueryStore interface {
	List(books []domain.Book) (Table, error)
	Query(books []domain.Book, where string) (Table, error)
}
