package localfs

import (
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/zawa-kyo/shelvia/internal/application"
)

func readBook(root, path string) (application.BookData, error) {
	var raw bookTOML
	if err := decodeTOML(root, path, &raw); err != nil {
		return application.BookData{}, err
	}
	return application.BookData{
		Path:       relativePath(root, path),
		Title:      raw.Title,
		Author:     raw.Author,
		Rating:     raw.Rating,
		ReadDate:   localDateTime(raw.ReadDate),
		Genre:      raw.Genre,
		Publisher:  raw.Publisher,
		Edition:    raw.Edition,
		Imprint:    raw.Imprint,
		Series:     raw.Series,
		Translator: raw.Translator,
		Summary:    raw.Thoughts.Summary,
		Body:       raw.Thoughts.Body,
	}, nil
}

func localDateTime(date toml.LocalDate) time.Time {
	if date.Year == 0 || date.Month == 0 || date.Day == 0 {
		return time.Time{}
	}
	return date.AsTime(time.UTC)
}

type bookTOML struct {
	Title      string         `toml:"title"`
	Author     string         `toml:"author"`
	Rating     int            `toml:"rating"`
	ReadDate   toml.LocalDate `toml:"read_date"`
	Genre      string         `toml:"genre"`
	Publisher  string         `toml:"publisher"`
	Edition    string         `toml:"edition"`
	Imprint    string         `toml:"imprint"`
	Series     string         `toml:"series"`
	Translator string         `toml:"translator"`
	Thoughts   thoughtsTOML   `toml:"thoughts"`
}

type thoughtsTOML struct {
	Summary string `toml:"summary"`
	Body    string `toml:"body"`
}
