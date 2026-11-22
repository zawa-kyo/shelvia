package queryengine

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zawa-kyo/shelvia/internal/domain"
)

func TestStoreList(t *testing.T) {
	t.Run("read_date descとtitle ascで書籍を返す", func(t *testing.T) {
		books := testBooks(t)

		result, err := Store{}.List(books)

		require.NoError(t, err)
		require.Equal(t, []string{"Alpha", "Beta", "Gamma"}, titles(result))
	})
}

func TestOrderArrayUsesValues(t *testing.T) {
	allowed, err := domain.NewAllowedValues(domain.AllowedValuesInput{
		Genres:     []string{"Novel"},
		Publishers: []string{"Example Publisher"},
		Imprints:   []string{"Example Paperback"},
		Editions:   []domain.EditionInput{{Name: "Paperback", ImprintRequired: true}},
	})
	require.NoError(t, err)
	books := []domain.Book{
		mustBook(t, allowed, "Letter", "A", 80, "2024-01-01", "Novel", ""),
		mustBook(t, allowed, "Symbol", "<", 80, "2024-01-01", "Novel", ""),
	}

	booksFound, err := (Store{}).Query(books, "rating >= 0 order by author")

	require.NoError(t, err)
	require.Equal(t, []string{"Symbol", "Letter"}, titles(booksFound))
}

func TestStoreQuery(t *testing.T) {
	tests := []struct {
		name  string
		where string
		want  []string
	}{
		{name: "数値比較で絞り込む", where: "rating >= 80", want: []string{"Alpha", "Gamma"}},
		{name: "未満で絞り込む", where: "rating < 80", want: []string{"Beta"}},
		{name: "以下で絞り込む", where: "rating <= 80", want: []string{"Beta", "Gamma"}},
		{name: "超過で絞り込む", where: "rating > 80", want: []string{"Alpha"}},
		{name: "不一致で絞り込む", where: `genre != "Novel"`, want: []string{"Beta"}},
		{name: "日付比較で絞り込む", where: `read_date = "2024-01-02"`, want: []string{"Alpha", "Beta"}},
		{name: "文字列一致で絞り込む", where: `genre = "Technical"`, want: []string{"Beta"}},
		{name: "likeで絞り込む", where: `title like "Al%"`, want: []string{"Alpha"}},
		{name: "andとorと括弧を解釈する", where: `(genre = "Novel" and rating >= 90) or title = "Beta"`, want: []string{"Alpha", "Beta"}},
		{name: "任意項目の未指定をnullとして扱う", where: "translator is null", want: []string{"Alpha", "Beta"}},
		{name: "任意項目の指定済みをnot nullとして扱う", where: "translator is not null", want: []string{"Gamma"}},
		{name: "nullは通常の不一致比較にも一致しない", where: `translator != "Translator"`, want: []string{}},
		{name: "nullとの比較はnotでも真にならない", where: `not translator = "Translator"`, want: []string{}},
		{name: "unknownとのandは真にならない", where: `translator = "Missing" and title = "Alpha"`, want: []string{}},
		{name: "unknownとのorは真の項だけに一致する", where: `translator = "Missing" or title = "Alpha"`, want: []string{"Alpha"}},
		{name: "notを解釈する", where: `not genre = "Technical"`, want: []string{"Alpha", "Gamma"}},
		{name: "タイトルを昇順で並び替える", where: "rating >= 70 order by title asc", want: []string{"Alpha", "Beta", "Gamma"}},
		{name: "レートを降順で並び替える", where: "rating >= 70 order by rating desc", want: []string{"Alpha", "Gamma", "Beta"}},
		{name: "order byのデフォルト値は昇順である", where: "rating >= 70 order by rating", want: []string{"Beta", "Gamma", "Alpha"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			books := testBooks(t)

			booksFound, err := Store{}.Query(books, tt.where)

			require.NoError(t, err)
			got := titles(booksFound)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestStoreQueryArrayAndLiteralSemantics(t *testing.T) {
	allowed, err := domain.NewAllowedValues(domain.AllowedValuesInput{
		Genres:     []string{"Novel", "Technical"},
		Publishers: []string{"Example Publisher"},
	})
	require.NoError(t, err)
	date := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	book, err := domain.NewBook(domain.BookDraft{
		Title:     "A' OR 1=1",
		Authors:   []string{"Alice", "Bob"},
		Genres:    []string{"Novel", "Technical"},
		Rating:    intPointer(90),
		ReadDate:  date,
		Publisher: "Example Publisher",
	}, allowed)
	require.NoError(t, err)

	for _, tc := range []struct {
		query string
		count int
	}{
		{`author = "Bob"`, 1},
		{`author != "Bob"`, 0},
		{`author != "Other"`, 1},
		{`genre like "Tech%"`, 1},
		{`title like "a%"`, 0},
		{`title = "A' OR 1=1"`, 1},
		{`translator = "Other" or title = "missing"`, 0},
	} {
		t.Run(tc.query, func(t *testing.T) {
			booksFound, err := (Store{}).Query([]domain.Book{book}, tc.query)

			require.NoError(t, err)
			require.Len(t, booksFound, tc.count)
		})
	}
}

func TestStoreQueryRejectsUnsupportedWhere(t *testing.T) {
	tests := []struct {
		name  string
		where string
	}{
		{name: "orderの後にbyが必要", where: "rating >= 80 order title"},
		{name: "order byの列名が必要", where: "rating >= 80 order by"},
		{name: "order byの未対応列は受け付けない", where: "rating >= 80 order by unknown"},
		{name: "order byの後に余分な構文は受け付けない", where: "rating >= 80 order by title desc limit 1"},
		{name: "select文は受け付けない", where: "select * from books"},
		{name: "in演算子は受け付けない", where: "title in ('A')"},
		{name: "二重等号演算子は受け付けない", where: "rating == 90"},
		{name: "関数呼び出しは受け付けない", where: "lower(title) = 'a'"},
		{name: "複数文は受け付けない", where: "title = 'A'; drop table books"},
		{name: "コメントは受け付けない", where: "title = 'A' -- comment"},
		{name: "更新文は受け付けない", where: "update books"},
		{name: "DDLは受け付けない", where: "drop table books"},
		{name: "joinは受け付けない", where: "title = 'A' join books"},
		{name: "unionは受け付けない", where: "title = 'A' union title = 'B'"},
		{name: "未対応の列は受け付けない", where: "unknown = 'A'"},
		{name: "文字列列と数値は比較できない", where: "title = 123"},
		{name: "文字列は引用符で囲む", where: "genre = Novel"},
		{name: "評価と文字列は比較できない", where: `rating = "90"`},
		{name: "読了日と数値は比較できない", where: `read_date = 2024`},
		{name: "読了日はISO形式だけ受け付ける", where: `read_date = "2024/01/02"`},
		{name: "評価にlikeは使えない", where: `rating like "9%"`},
		{name: "読了日にlikeは使えない", where: `read_date like "2024%"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			books := testBooks(t)

			_, err := Store{}.Query(books, tt.where)

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

func mustBook(
	t *testing.T,
	allowed domain.AllowedValues,
	title, author string,
	rating int,
	date, genre, translator string,
) domain.Book {
	t.Helper()
	readDate, err := time.Parse("2006-01-02", date)
	require.NoError(t, err)

	translators := []string{}
	if translator != "" {
		translators = append(translators, translator)
	}
	book, err := domain.NewBook(domain.BookDraft{
		Translators: translators,
		Title:       title,
		Authors:     []string{author},
		Rating:      intPointer(rating),
		ReadDate:    readDate,
		Genres:      []string{genre},
		Publisher:   "Example Publisher",
		Edition:     "Paperback",
		Imprint:     "Example Paperback",
	}, allowed)
	require.NoError(t, err)

	return book
}

func intPointer(value int) *int {
	return &value
}

func titles(books []domain.Book) []string {
	got := make([]string, len(books))
	for i, book := range books {
		got[i] = book.Title().String()
	}
	return got
}
