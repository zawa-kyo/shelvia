package application

import "fmt"

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

		result, err := service.query.List(books)
		return Output{Kind: List, Books: result}, err
	case Search:
		books, err := service.loadBooks(command.ShelfRoot)
		if err != nil {
			return Output{}, err
		}

		result, err := service.query.Query(books, command.Where)
		return Output{Kind: Search, Books: result}, err
	case Show:
		books, err := service.loadBooks(command.ShelfRoot)
		if err != nil {
			return Output{}, err
		}

		book, err := findBookByTitle(books, command.Title)
		if err != nil {
			return Output{}, err
		}
		return Output{Kind: Show, Book: book}, nil
	case Export:
		books, err := service.loadBooks(command.ShelfRoot)
		if err != nil {
			return Output{}, err
		}
		return Output{Kind: Export, Books: books}, nil
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

func plural(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
