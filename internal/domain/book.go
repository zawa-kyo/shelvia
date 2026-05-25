package domain

import (
	"fmt"
)

// Aggregate root for a reading-log entry.
type Book struct {
	title      Title
	author     Author
	rating     Rating
	readDate   ReadDate
	genre      Genre
	publisher  Publisher
	edition    Edition
	imprint    Imprint
	series     OptionalText
	translator OptionalText
	summary    OptionalText
	body       OptionalText
	filePath   FilePath
}

// Creates a book and validates all domain invariants.
func NewBook(draft BookDraft, allowed AllowedValues) (Book, error) {
	title, err := NewTitle(draft.Title)
	if err != nil {
		return Book{}, err
	}
	author, err := NewAuthor(draft.Author)
	if err != nil {
		return Book{}, err
	}
	rating, err := NewRating(draft.Rating)
	if err != nil {
		return Book{}, err
	}
	readDate, err := NewReadDate(draft.ReadDate)
	if err != nil {
		return Book{}, err
	}
	genre, err := NewGenre(draft.Genre)
	if err != nil {
		return Book{}, err
	}
	if !allowed.ContainsGenre(genre) {
		return Book{}, validationError("genre", fmt.Sprintf("unknown genre: %s", genre.String()))
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
		title:      title,
		author:     author,
		rating:     rating,
		readDate:   readDate,
		genre:      genre,
		publisher:  publisher,
		edition:    edition,
		imprint:    imprint,
		series:     NewOptionalText(draft.Series),
		translator: NewOptionalText(draft.Translator),
		summary:    NewOptionalText(draft.Summary),
		body:       NewOptionalText(draft.Body),
		filePath:   NewFilePath(draft.FilePath),
	}, nil
}

// Returns the book title.
func (book Book) Title() Title {
	return book.title
}

// Returns the book author.
func (book Book) Author() Author {
	return book.author
}

// Returns the book rating.
func (book Book) Rating() Rating {
	return book.rating
}

// Returns the date when the book was read.
func (book Book) ReadDate() ReadDate {
	return book.readDate
}

// Returns the book genre.
func (book Book) Genre() Genre {
	return book.genre
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

// Returns the optional translator text.
func (book Book) Translator() OptionalText {
	return book.translator
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
