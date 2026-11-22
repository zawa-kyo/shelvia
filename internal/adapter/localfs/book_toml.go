package localfs

import (
	"fmt"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/zawa-kyo/shelvia/internal/application"
)

func readBook(root, path string) (application.BookData, error) {
	var raw bookTOML
	if err := decodeTOML(root, path, &raw); err != nil {
		return application.BookData{}, err
	}

	authors, err := stringArray(raw.Author)
	if err != nil {
		return application.BookData{}, fmt.Errorf("%s: author: %w", relativePath(root, path), err)
	}

	genres, err := stringArray(raw.Genre)
	if err != nil {
		return application.BookData{}, fmt.Errorf("%s: genre: %w", relativePath(root, path), err)
	}

	translators, err := stringArray(raw.Translator)
	if err != nil {
		return application.BookData{}, fmt.Errorf("%s: translator: %w", relativePath(root, path), err)
	}

	return application.BookData{
		Authors:     authors,
		Genres:      genres,
		Translators: translators,
		Path:        relativePath(root, path),
		Title:       raw.Title,
		Rating:      raw.Rating,
		ReadDate:    localDateTime(raw.ReadDate),
		Publisher:   raw.Publisher,
		Edition:     raw.Edition,
		Imprint:     raw.Imprint,
		Series:      raw.Series,
		Summary:     raw.Thoughts.Summary,
		Body:        raw.Thoughts.Body,
	}, nil
}

func stringArray(value any) ([]string, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case string:
		if strings.TrimSpace(v) == "" {
			return []string{}, nil
		}
		return []string{v}, nil
	case []any:
		values := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("expected an array of strings")
			}
			values = append(values, s)
		}
		return values, nil
	default:
		return nil, fmt.Errorf("expected a string or an array of strings")
	}
}

func localDateTime(date toml.LocalDate) time.Time {
	if date.Year == 0 || date.Month == 0 || date.Day == 0 {
		return time.Time{}
	}
	return date.AsTime(time.UTC)
}

type bookTOML struct {
	Title      string         `toml:"title"`
	Author     any            `toml:"author"`
	Rating     *int           `toml:"rating"`
	ReadDate   toml.LocalDate `toml:"read_date"`
	Genre      any            `toml:"genre"`
	Publisher  string         `toml:"publisher"`
	Edition    string         `toml:"edition"`
	Imprint    string         `toml:"imprint"`
	Series     string         `toml:"series"`
	Translator any            `toml:"translator"`
	Thoughts   thoughtsTOML   `toml:"thoughts"`
}

type thoughtsTOML struct {
	Summary string `toml:"summary"`
	Body    string `toml:"body"`
}
