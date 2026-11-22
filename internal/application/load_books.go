package application

import (
	"fmt"

	"github.com/zawa-kyo/shelvia/internal/domain"
)

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
		Authors:     data.Authors,
		Genres:      data.Genres,
		Translators: data.Translators,
		Title:       data.Title,

		Rating:   data.Rating,
		ReadDate: data.ReadDate,

		Publisher: data.Publisher,
		Edition:   data.Edition,
		Imprint:   data.Imprint,
		Series:    data.Series,

		Summary:  data.Summary,
		Body:     data.Body,
		FilePath: data.Path,
	}
}
