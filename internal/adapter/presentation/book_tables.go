package presentation

import (
	"encoding/json"
	"strconv"

	"github.com/zawa-kyo/shelvia/internal/application"
	"github.com/zawa-kyo/shelvia/internal/domain"
)

type tableData struct {
	Headers []string
	Rows    [][]string
}

func outputTable(output application.Output) tableData {
	switch output.Kind {
	case application.List, application.Search:
		return bookListTable(output.Books)
	case application.Show:
		return bookDetailsTable(output.Book)
	case application.Export:
		return exportTable(output.Books)
	default:
		return tableData{}
	}
}

func bookListTable(books []domain.Book) tableData {
	table := tableData{Headers: []string{
		"read_date",
		"rating",
		"title",
		"author",
		"genre",
		"publisher",
	}}

	for _, book := range books {
		table.Rows = append(table.Rows, []string{
			book.ReadDate().String(),
			strconv.Itoa(book.Rating().Int()),
			book.Title().String(),
			arrayJSON(book.Authors()),
			arrayJSON(book.Genres()),
			book.Publisher().String(),
		})
	}

	return table
}

func exportTable(books []domain.Book) tableData {
	table := tableData{Headers: []string{
		"title",
		"author",
		"rating",
		"read_date",
		"genre",
		"publisher",
		"edition",
		"imprint",
		"series",
		"translator",
		"thoughts.summary",
		"thoughts.body",
	}}

	for _, book := range books {
		table.Rows = append(table.Rows, []string{
			book.Title().String(),
			arrayJSON(book.Authors()),
			strconv.Itoa(book.Rating().Int()),
			book.ReadDate().String(),
			arrayJSON(book.Genres()),
			book.Publisher().String(),
			book.Edition().String(),
			book.Imprint().String(),
			book.Series().String(),
			arrayJSON(book.Translators()),
			book.Summary().String(),
			book.Body().String(),
		})
	}

	return table
}

func arrayJSON(values []string) string {
	if values == nil {
		values = []string{}
	}
	data, _ := json.Marshal(values)
	return string(data)
}

func bookDetailsTable(book domain.Book) tableData {
	rows := [][]string{
		{"path", book.FilePath().String()},
		{"title", book.Title().String()},
		{"author", arrayJSON(book.Authors())},
		{"rating", strconv.Itoa(book.Rating().Int())},
		{"read_date", book.ReadDate().String()},
		{"genre", arrayJSON(book.Genres())},
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
	if values := book.Translators(); len(values) != 0 {
		rows = append(rows, []string{"translator", arrayJSON(values)})
	}
	if value := book.Summary().String(); value != "" {
		rows = append(rows, []string{"summary", value})
	}
	if value := book.Body().String(); value != "" {
		rows = append(rows, []string{"body", value})
	}

	return tableData{Headers: []string{"field", "value"}, Rows: rows}
}
