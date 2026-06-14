package application

import (
	"fmt"
	"strings"

	"github.com/zawa-kyo/shelvia/internal/domain"
)

// Coordinates use cases across shelf storage, domain validation, and querying.
type Service struct {
	shelf ShelfRepository
	query QueryStore
}

// Creates an application service from external ports.
func NewService(shelf ShelfRepository, query QueryStore) Service {
	return Service{shelf: shelf, query: query}
}

// Runs a command and returns a presentation-neutral result.
func (service Service) Run(command Command) (Output, error) {
	switch command.Kind {
	case Init:
		if err := service.shelf.Init(command.ShelfRoot); err != nil {
			return Output{}, err
		}
		return Output{Message: "Initialized shelf."}, nil
	case Add:
		path, err := service.shelf.CreateBook(command.ShelfRoot, command.Title, command.ReadDate)
		if err != nil {
			return Output{}, err
		}
		return Output{Message: fmt.Sprintf("Created %s.", path)}, nil
	case Validate:
		books, err := service.loadBooks(command.ShelfRoot)
		if err != nil {
			return Output{}, err
		}
		return Output{Message: fmt.Sprintf("Validated %d book%s, 1 config file.", len(books), plural(len(books)))}, nil
	case List:
		books, err := service.loadBooks(command.ShelfRoot)
		if err != nil {
			return Output{}, err
		}
		table, err := service.query.List(books)
		return Output{Table: table}, err
	case Search:
		books, err := service.loadBooks(command.ShelfRoot)
		if err != nil {
			return Output{}, err
		}
		table, err := service.query.Query(books, command.Where)
		return Output{Table: table}, err
	case Show:
		books, err := service.loadBooks(command.ShelfRoot)
		if err != nil {
			return Output{}, err
		}
		book, err := findBookByTitle(books, command.Title)
		if err != nil {
			return Output{}, err
		}
		return Output{Table: bookDetailsTable(book)}, nil
	case Path:
		books, err := service.loadBooks(command.ShelfRoot)
		if err != nil {
			return Output{}, err
		}
		book, err := findBookByTitle(books, command.Title)
		if err != nil {
			return Output{}, err
		}
		return Output{Message: book.FilePath().String()}, nil
	default:
		return Output{}, fmt.Errorf("unknown command")
	}
}

func (service Service) loadBooks(root string) ([]domain.Book, error) {
	data, err := service.shelf.Load(root)
	if err != nil {
		return nil, err
	}
	allowed, err := domain.NewAllowedValues(toAllowedValuesInput(data.Config))
	if err != nil {
		return nil, fmt.Errorf("config.toml: %w", err)
	}

	books := make([]domain.Book, 0, len(data.Books))
	for _, bookData := range data.Books {
		book, err := domain.NewBook(toBookDraft(bookData), allowed)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", bookData.Path, err)
		}
		books = append(books, book)
	}
	return books, nil
}

func toAllowedValuesInput(data AllowedValuesData) domain.AllowedValuesInput {
	editions := make([]domain.EditionInput, len(data.Editions))
	for i, edition := range data.Editions {
		editions[i] = domain.EditionInput{
			Name:            edition.Name,
			ImprintRequired: edition.ImprintRequired,
		}
	}
	return domain.AllowedValuesInput{
		Genres:     data.Genres,
		Publishers: data.Publishers,
		Imprints:   data.Imprints,
		Editions:   editions,
	}
}

func toBookDraft(data BookData) domain.BookDraft {
	return domain.BookDraft{
		Title:      data.Title,
		Author:     data.Author,
		Rating:     data.Rating,
		ReadDate:   data.ReadDate,
		Genre:      data.Genre,
		Publisher:  data.Publisher,
		Edition:    data.Edition,
		Imprint:    data.Imprint,
		Series:     data.Series,
		Translator: data.Translator,
		Summary:    data.Summary,
		Body:       data.Body,
		FilePath:   data.Path,
	}
}

func plural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func findBookByTitle(books []domain.Book, title string) (domain.Book, error) {
	matches := make([]domain.Book, 0, 1)
	for _, book := range books {
		if book.Title().String() == title {
			matches = append(matches, book)
		}
	}
	if len(matches) == 0 {
		return domain.Book{}, fmt.Errorf("book title not found: %s", title)
	}
	if len(matches) > 1 {
		paths := make([]string, len(matches))
		for i, book := range matches {
			paths[i] = book.FilePath().String()
		}
		return domain.Book{}, fmt.Errorf("book title is ambiguous: %s (%s)", title, strings.Join(paths, ", "))
	}
	return matches[0], nil
}

func bookDetailsTable(book domain.Book) Table {
	rows := [][]string{
		{"path", book.FilePath().String()},
		{"title", book.Title().String()},
		{"author", book.Author().String()},
		{"rating", fmt.Sprintf("%d", book.Rating().Int())},
		{"read_date", book.ReadDate().String()},
		{"genre", book.Genre().String()},
		{"publisher", book.Publisher().String()},
	}
	if value := book.Edition().String(); value != "" {
		rows = append(rows, []string{"edition", value})
	}
	if value := book.Imprint().String(); value != "" {
		rows = append(rows, []string{"imprint", value})
	}
	if value := book.Series().String(); value != "" {
		rows = append(rows, []string{"series", value})
	}
	if value := book.Translator().String(); value != "" {
		rows = append(rows, []string{"translator", value})
	}
	if value := book.Summary().String(); value != "" {
		rows = append(rows, []string{"summary", value})
	}
	if value := book.Body().String(); value != "" {
		rows = append(rows, []string{"body", value})
	}
	return Table{Headers: []string{"field", "value"}, Rows: rows}
}
