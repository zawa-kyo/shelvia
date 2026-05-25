package queryengine

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/zawa-kyo/shelvia/internal/application"
	"github.com/zawa-kyo/shelvia/internal/domain"
)

var headers = []string{"read_date", "rating", "title", "author", "genre", "publisher"}

// Store provides table queries over validated books.
type Store struct{}

// Lists books in the default order.
func (Store) List(books []domain.Book) (application.Table, error) {
	return tableFromBooks(sortBooks(books)), nil
}

// Filters books with a restricted SQL-like where fragment.
func (Store) Query(books []domain.Book, where string) (application.Table, error) {
	expr, err := parseWhere(where)
	if err != nil {
		return application.Table{}, err
	}
	filtered := make([]domain.Book, 0, len(books))
	for _, book := range books {
		ok, err := expr.eval(rowFromBook(book))
		if err != nil {
			return application.Table{}, err
		}
		if ok {
			filtered = append(filtered, book)
		}
	}
	return tableFromBooks(sortBooks(filtered)), nil
}

func sortBooks(books []domain.Book) []domain.Book {
	sorted := append([]domain.Book(nil), books...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left, right := sorted[i], sorted[j]
		if !left.ReadDate().Time().Equal(right.ReadDate().Time()) {
			return left.ReadDate().Time().After(right.ReadDate().Time())
		}
		return left.Title().String() < right.Title().String()
	})
	return sorted
}

func tableFromBooks(books []domain.Book) application.Table {
	rows := make([][]string, len(books))
	for i, book := range books {
		rows[i] = []string{
			book.ReadDate().String(),
			strconv.Itoa(book.Rating().Int()),
			book.Title().String(),
			book.Author().String(),
			book.Genre().String(),
			book.Publisher().String(),
		}
	}
	return application.Table{Headers: headers, Rows: rows}
}

func rowFromBook(book domain.Book) map[string]string {
	row := map[string]string{
		"title":      book.Title().String(),
		"author":     book.Author().String(),
		"rating":     strconv.Itoa(book.Rating().Int()),
		"read_date":  book.ReadDate().String(),
		"genre":      book.Genre().String(),
		"publisher":  book.Publisher().String(),
		"edition":    optionalControlled(book.Edition().IsSpecified(), book.Edition().String()),
		"imprint":    optionalControlled(book.Imprint().IsSpecified(), book.Imprint().String()),
		"series":     optionalText(book.Series()),
		"translator": optionalText(book.Translator()),
	}
	return row
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

type expr interface {
	eval(row map[string]string) (bool, error)
}

type comparison struct {
	column string
	op     string
	value  string
}

func (node comparison) eval(row map[string]string) (bool, error) {
	left := row[node.column]
	if node.op == "is null" {
		return left == "", nil
	}
	if node.op == "is not null" {
		return left != "", nil
	}
	if left == "" {
		return false, nil
	}
	if node.column == "rating" {
		leftNumber, err := strconv.Atoi(left)
		if err != nil {
			return false, err
		}
		rightNumber, err := strconv.Atoi(node.value)
		if err != nil {
			return false, fmt.Errorf("rating comparison requires a number")
		}
		return compareInt(leftNumber, rightNumber, node.op), nil
	}
	if node.op == "like" {
		return like(left, node.value), nil
	}
	return compareString(left, node.value, node.op), nil
}

type logical struct {
	op          string
	left, right expr
}

func (node logical) eval(row map[string]string) (bool, error) {
	left, err := node.left.eval(row)
	if err != nil {
		return false, err
	}
	if node.op == "and" && !left {
		return false, nil
	}
	if node.op == "or" && left {
		return true, nil
	}
	right, err := node.right.eval(row)
	if err != nil {
		return false, err
	}
	if node.op == "and" {
		return left && right, nil
	}
	return left || right, nil
}

type negation struct {
	inner expr
}

func (node negation) eval(row map[string]string) (bool, error) {
	value, err := node.inner.eval(row)
	return !value, err
}

func compareInt(left, right int, op string) bool {
	switch op {
	case "=":
		return left == right
	case "!=":
		return left != right
	case "<":
		return left < right
	case "<=":
		return left <= right
	case ">":
		return left > right
	case ">=":
		return left >= right
	default:
		return false
	}
}

func compareString(left, right, op string) bool {
	switch op {
	case "=":
		return left == right
	case "!=":
		return left != right
	case "<":
		return left < right
	case "<=":
		return left <= right
	case ">":
		return left > right
	case ">=":
		return left >= right
	default:
		return false
	}
}

func like(value, pattern string) bool {
	quoted := regexp.QuoteMeta(pattern)
	quoted = strings.ReplaceAll(quoted, "%", ".*")
	quoted = strings.ReplaceAll(quoted, "_", ".")
	return regexp.MustCompile("^" + quoted + "$").MatchString(value)
}
