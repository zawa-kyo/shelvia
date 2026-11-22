package queryengine

import (
	"database/sql"
	"encoding/json"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/zawa-kyo/shelvia/internal/application"
	"github.com/zawa-kyo/shelvia/internal/domain"

	_ "modernc.org/sqlite"
)

var headers = []string{"read_date", "rating", "title", "author", "genre", "publisher"}

// Store provides table queries over validated books.
type Store struct{}

// Lists books in the default order.
func (Store) List(books []domain.Book) (application.Table, error) {
	return queryBooks(books, queryFragment{})
}

// Filters books with a restricted SQL-like where fragment.
func (Store) Query(books []domain.Book, where string) (application.Table, error) {
	query, err := parseQueryFragment(where)
	if err != nil {
		return application.Table{}, err
	}
	return queryBooks(books, query)
}

func queryBooks(books []domain.Book, query queryFragment) (application.Table, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return application.Table{}, err
	}
	defer db.Close()
	// An in-memory SQLite database belongs to a single connection.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA case_sensitive_like = ON`); err != nil {
		return application.Table{}, err
	}
	if _, err := db.Exec(`CREATE TABLE books (id INTEGER PRIMARY KEY, title TEXT, rating INTEGER, read_date TEXT, publisher TEXT, edition TEXT, imprint TEXT, series TEXT, author TEXT, genre TEXT, translator TEXT)`); err != nil {
		return application.Table{}, err
	}
	tx, err := db.Begin()
	if err != nil {
		return application.Table{}, err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`INSERT INTO books VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return application.Table{}, err
	}
	defer stmt.Close()
	for i, book := range books {
		row := rowFromBook(book)
		if _, err := stmt.Exec(i, book.Title().String(), book.Rating().Int(), book.ReadDate().String(), book.Publisher().String(), nullIfEmpty(row.scalars["edition"]), nullIfEmpty(row.scalars["imprint"]), nullIfEmpty(row.scalars["series"]), jsonArray(book.Authors()), jsonArray(book.Genres()), jsonArray(book.Translators())); err != nil {
			return application.Table{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return application.Table{}, err
	}
	condition := "1"
	var args []any
	if query.where != nil {
		condition, args = query.where.sql()
	}
	rows, err := db.Query("SELECT id FROM books WHERE "+condition, args...)
	if err != nil {
		return application.Table{}, err
	}
	defer rows.Close()
	filtered := make([]domain.Book, 0, len(books))
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return application.Table{}, err
		}
		filtered = append(filtered, books[id])
	}
	if err := rows.Err(); err != nil {
		return application.Table{}, err
	}
	return tableFromBooks(sortQueriedBooks(filtered, query.order)), nil
}

func sortBooks(books []domain.Book) []domain.Book {
	sorted := append([]domain.Book(nil), books...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return compareDefault(sorted[i], sorted[j]) < 0
	})
	return sorted
}

func sortQueriedBooks(books []domain.Book, order orderSpec) []domain.Book {
	if order.column == "" {
		return sortBooks(books)
	}
	sorted := append([]domain.Book(nil), books...)
	sort.SliceStable(sorted, func(i, j int) bool {
		cmp := compareByColumn(sorted[i], sorted[j], order.column)
		if cmp == 0 {
			return compareDefault(sorted[i], sorted[j]) < 0
		}
		if order.desc {
			return cmp > 0
		}
		return cmp < 0
	})
	return sorted
}

func compareDefault(left, right domain.Book) int {
	if !left.ReadDate().Time().Equal(right.ReadDate().Time()) {
		if left.ReadDate().Time().After(right.ReadDate().Time()) {
			return -1
		}
		return 1
	}
	return strings.Compare(left.Title().String(), right.Title().String())
}

func compareByColumn(left, right domain.Book, column string) int {
	leftRow, rightRow := rowFromBook(left), rowFromBook(right)
	switch column {
	case "rating":
		return left.Rating().Int() - right.Rating().Int()
	case "read_date":
		return left.ReadDate().Time().Compare(right.ReadDate().Time())
	case "author", "genre", "translator":
		return slices.Compare(leftRow.arrays[column], rightRow.arrays[column])
	default:
		return strings.Compare(leftRow.scalars[column], rightRow.scalars[column])
	}
}

func tableFromBooks(books []domain.Book) application.Table {
	rows := make([][]string, len(books))
	for i, book := range books {
		rows[i] = []string{
			book.ReadDate().String(),
			strconv.Itoa(book.Rating().Int()),
			book.Title().String(),
			jsonArray(book.Authors()),
			jsonArray(book.Genres()),
			book.Publisher().String(),
		}
	}
	return application.Table{Headers: headers, Rows: rows}
}

func jsonArray(values []string) string {
	if values == nil {
		values = []string{}
	}
	data, _ := json.Marshal(values)
	return string(data)
}

type queryRow struct {
	scalars map[string]string
	arrays  map[string][]string
}

func rowFromBook(book domain.Book) queryRow {
	return queryRow{
		scalars: map[string]string{
			"title": book.Title().String(), "rating": strconv.Itoa(book.Rating().Int()),
			"read_date": book.ReadDate().String(), "publisher": book.Publisher().String(),
			"edition": optionalControlled(book.Edition().IsSpecified(), book.Edition().String()),
			"imprint": optionalControlled(book.Imprint().IsSpecified(), book.Imprint().String()),
			"series":  optionalText(book.Series()),
		},
		arrays: map[string][]string{"author": book.Authors(), "genre": book.Genres(), "translator": book.Translators()},
	}
}

func optionalControlled(specified bool, value string) string {
	if !specified {
		return ""
	}
	return value
}

func optionalText(text domain.OptionalText) string {
	if !text.IsSpecified() {
		return ""
	}
	return text.String()
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
