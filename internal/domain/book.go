package domain

import (
	"fmt"
	"strings"
)

// Aggregate root for a reading-log entry.
type Book struct {
	authors     []string
	genres      []string
	translators []string
	title       Title
	rating      Rating
	readDate    ReadDate
	publisher   Publisher
	edition     Edition
	imprint     Imprint
	series      OptionalText
	summary     OptionalText
	body        OptionalText
	filePath    FilePath
}

// Creates a book and validates all domain invariants.
func NewBook(draft BookDraft, allowed AllowedValues) (Book, error) {
	if len(draft.Authors) == 0 {
		return Book{}, validationError("author", "required")
	}
	authors := make([]string, 0, len(draft.Authors))
	for _, value := range draft.Authors {
		author, err := NewAuthor(value)
		if err != nil {
			return Book{}, err
		}
		authors = append(authors, author.String())
	}

	if len(draft.Genres) == 0 {
		return Book{}, validationError("genre", "required")
	}
	genres := make([]string, 0, len(draft.Genres))
	for _, value := range draft.Genres {
		genre, err := NewGenre(value)
		if err != nil {
			return Book{}, err
		}
		if !allowed.ContainsGenre(genre) {
			return Book{}, validationError("genre", fmt.Sprintf("unknown genre: %s", genre.String()))
		}
		genres = append(genres, genre.String())
	}

	translators := make([]string, 0, len(draft.Translators))
	for _, value := range draft.Translators {
		value = strings.TrimSpace(value)
		if value == "" {
			return Book{}, validationError("translator", "must not be empty")
		}
		translators = append(translators, value)
	}

	title, err := NewTitle(draft.Title)
	if err != nil {
		return Book{}, err
	}

	if draft.Rating == nil {
		return Book{}, validationError("rating", "required")
	}

	rating, err := NewRating(*draft.Rating)
	if err != nil {
		return Book{}, err
	}

	readDate, err := NewReadDate(draft.ReadDate)
	if err != nil {
		return Book{}, err
	}

	publisher, err := NewPublisher(draft.Publisher)
	if err != nil {
		return Book{}, err
	}
	if !allowed.ContainsPublisher(publisher) {
		return Book{}, validationError("publisher", fmt.Sprintf("unknown publisher: %s", publisher.String()))
	}

	edition, err := NewEdition(draft.Edition)
	if err != nil {
		return Book{}, err
	}
	var editionRule EditionRule
	if edition.IsSpecified() {
		var ok bool
		editionRule, ok = allowed.FindEdition(edition)
		if !ok {
			return Book{}, validationError("edition", fmt.Sprintf("unknown edition: %s", edition.String()))
		}
	}

	imprint, err := NewImprint(draft.Imprint)
	if err != nil {
		return Book{}, err
	}
	if imprint.IsSpecified() && !edition.IsSpecified() {
		return Book{}, validationError("imprint", fmt.Sprintf("imprint requires edition: %s", imprint.String()))
	}
	if edition.IsSpecified() && editionRule.ImprintRequired() && !imprint.IsSpecified() {
		return Book{}, validationError("imprint", fmt.Sprintf("imprint is required for edition: %s", edition.String()))
	}
	if imprint.IsSpecified() && !allowed.ContainsImprint(imprint) {
		return Book{}, validationError("imprint", fmt.Sprintf("unknown imprint: %s", imprint.String()))
	}

	return Book{
		authors:     authors,
		genres:      genres,
		translators: translators,
		title:       title,
		rating:      rating,
		readDate:    readDate,
		publisher:   publisher,
		edition:     edition,
		imprint:     imprint,
		series:      NewOptionalText(draft.Series),
		summary:     NewOptionalText(draft.Summary),
		body:        NewOptionalText(draft.Body),
		filePath:    NewFilePath(draft.FilePath),
	}, nil
}

func (book Book) Authors() []string     { return append([]string(nil), book.authors...) }
func (book Book) Genres() []string      { return append([]string(nil), book.genres...) }
func (book Book) Translators() []string { return append([]string(nil), book.translators...) }

// Returns the book title.
func (book Book) Title() Title {
	return book.title
}

// Returns the book rating.
func (book Book) Rating() Rating {
	return book.rating
}

// Returns the date when the book was read.
func (book Book) ReadDate() ReadDate {
	return book.readDate
}

// Returns the book publisher.
func (book Book) Publisher() Publisher {
	return book.publisher
}

// Returns the optional book edition.
func (book Book) Edition() Edition {
	return book.edition
}

// Returns the optional book imprint.
func (book Book) Imprint() Imprint {
	return book.imprint
}

// Returns the optional series text.
func (book Book) Series() OptionalText {
	return book.series
}

// Returns the optional short thoughts text.
func (book Book) Summary() OptionalText {
	return book.summary
}

// Returns the optional long thoughts text.
func (book Book) Body() OptionalText {
	return book.body
}

// Returns the source file path used for diagnostics.
func (book Book) FilePath() FilePath {
	return book.filePath
}
