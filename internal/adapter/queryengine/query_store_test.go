package queryengine

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zawa-kyo/shelvia/internal/domain"
)

func TestStoreList(t *testing.T) {
	t.Run("read_date descとtitle ascで表示列を返す", func(t *testing.T) {
		books := testBooks(t)
		wantRows := [][]string{
			{"2024-01-02", "90", "Alpha", "Author A", "Novel", "Example Publisher"},
			{"2024-01-02", "70", "Beta", "Author B", "Technical", "Example Publisher"},
			{"2024-01-01", "80", "Gamma", "Author C", "Novel", "Example Publisher"},
		}

		table, err := Store{}.List(books)

		require.NoError(t, err)
		require.Equal(t, wantRows, table.Rows)
	})
}

func TestStoreQuery(t *testing.T) {
	tests := []struct {
		name  string
		where string
		want  []string
	}{
		{name: "数値比較で絞り込む", where: "rating >= 80", want: []string{"Alpha", "Gamma"}},
		{name: "日付比較で絞り込む", where: `read_date = "2024-01-02"`, want: []string{"Alpha", "Beta"}},
		{name: "文字列一致で絞り込む", where: `genre = "Technical"`, want: []string{"Beta"}},
		{name: "likeで絞り込む", where: `title like "Al%"`, want: []string{"Alpha"}},
		{name: "andとorと括弧を解釈する", where: `(genre = "Novel" and rating >= 90) or title = "Beta"`, want: []string{"Alpha", "Beta"}},
		{name: "任意項目の未指定をnullとして扱う", where: "translator is null", want: []string{"Alpha", "Beta"}},
		{name: "notを解釈する", where: `not genre = "Technical"`, want: []string{"Alpha", "Gamma"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			books := testBooks(t)

			table, err := Store{}.Query(books, tt.where)

			require.NoError(t, err)
			got := titles(table.Rows)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestStoreQueryRejectsUnsupportedWhere(t *testing.T) {
	tests := []string{
		"rating >= 80 order by title",
		"select * from books",
		"title in ('A')",
		"lower(title) = 'a'",
		"title = 'A'; drop table books",
		"title = 'A' -- comment",
		"unknown = 'A'",
		"title = 123",
		"genre = Novel",
		`rating = "90"`,
		`read_date = 2024`,
		`read_date = "2024/01/02"`,
		`rating like "9%"`,
		`read_date like "2024%"`,
	}

	for _, where := range tests {
		t.Run(where, func(t *testing.T) {
			books := testBooks(t)

			_, err := Store{}.Query(books, where)

			require.Error(t, err)
			require.NotEmpty(t, strings.TrimSpace(err.Error()))
		})
	}
}

func testBooks(t *testing.T) []domain.Book {
	t.Helper()
	allowed, err := domain.NewAllowedValues(domain.AllowedValuesInput{
		Genres:     []string{"Novel", "Technical"},
		Publishers: []string{"Example Publisher"},
		Imprints:   []string{"Example Paperback"},
		Editions: []domain.EditionInput{
			{Name: "Paperback", ImprintRequired: true},
		},
	})
	require.NoError(t, err)

	return []domain.Book{
		mustBook(t, allowed, "Gamma", "Author C", 80, "2024-01-01", "Novel", "Translator"),
		mustBook(t, allowed, "Beta", "Author B", 70, "2024-01-02", "Technical", ""),
		mustBook(t, allowed, "Alpha", "Author A", 90, "2024-01-02", "Novel", ""),
	}
}

func mustBook(t *testing.T, allowed domain.AllowedValues, title, author string, rating int, date string, genre string, translator string) domain.Book {
	t.Helper()
	readDate, err := time.Parse("2006-01-02", date)
	require.NoError(t, err)

	book, err := domain.NewBook(domain.BookDraft{
		Title:      title,
		Author:     author,
		Rating:     rating,
		ReadDate:   readDate,
		Genre:      genre,
		Publisher:  "Example Publisher",
		Edition:    "Paperback",
		Imprint:    "Example Paperback",
		Translator: translator,
	}, allowed)
	require.NoError(t, err)

	return book
}

func titles(rows [][]string) []string {
	got := make([]string, len(rows))
	for i, row := range rows {
		got[i] = row[2]
	}
	return got
}
