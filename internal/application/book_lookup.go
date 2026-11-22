package application

import (
	"fmt"
	"strings"

	"github.com/zawa-kyo/shelvia/internal/domain"
)

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
