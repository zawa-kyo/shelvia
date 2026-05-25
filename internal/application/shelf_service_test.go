package application

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/zawa-kyo/shelvia/internal/domain"
)

func TestServiceRunValidate(t *testing.T) {
	t.Run("shelfを読み込んで検証件数を返す", func(t *testing.T) {
		service := NewService(fakeShelf{data: validShelfData()}, fakeQuery{})

		output, err := service.Run(Command{Kind: Validate, ShelfRoot: "shelf"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.Message != "Validated 2 books, 1 config file." {
			t.Fatalf("message = %q", output.Message)
		}
	})

	t.Run("domain validation errorにはファイルパスを付ける", func(t *testing.T) {
		data := validShelfData()
		data.Books[0].Genre = "Unknown"
		service := NewService(fakeShelf{data: data}, fakeQuery{})

		_, err := service.Run(Command{Kind: Validate, ShelfRoot: "shelf"})
		if err == nil {
			t.Fatal("expected error")
		}
		if got := err.Error(); got != "B.toml: genre: unknown genre: Unknown" {
			t.Fatalf("error = %q", got)
		}
	})
}

func TestServiceRunListAndQuery(t *testing.T) {
	t.Run("listは検証済みbookをquery portへ渡す", func(t *testing.T) {
		query := &recordingQuery{table: Table{Headers: []string{"title"}, Rows: [][]string{{"A"}}}}
		service := NewService(fakeShelf{data: validShelfData()}, query)

		output, err := service.Run(Command{Kind: List, ShelfRoot: "shelf"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(output.Table, query.table) {
			t.Fatalf("table = %#v, want %#v", output.Table, query.table)
		}
		if query.listCount != 1 {
			t.Fatalf("list calls = %d, want 1", query.listCount)
		}
		if len(query.books) != 2 {
			t.Fatalf("books = %d, want 2", len(query.books))
		}
	})

	t.Run("queryはwhere fragmentをquery portへ渡す", func(t *testing.T) {
		query := &recordingQuery{table: Table{Headers: []string{"title"}, Rows: [][]string{{"B"}}}}
		service := NewService(fakeShelf{data: validShelfData()}, query)

		_, err := service.Run(Command{Kind: Query, ShelfRoot: "shelf", Where: `rating >= 90`})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if query.where != `rating >= 90` {
			t.Fatalf("where = %q", query.where)
		}
	})
}

func TestServiceRunInitAndNew(t *testing.T) {
	t.Run("initはrepositoryへ委譲する", func(t *testing.T) {
		shelf := &recordingShelf{newPath: "Some Book.toml"}
		service := NewService(shelf, fakeQuery{})

		_, err := service.Run(Command{Kind: Init, ShelfRoot: "shelf"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if shelf.initRoot != "shelf" {
			t.Fatalf("init root = %q", shelf.initRoot)
		}
	})

	t.Run("newはrepositoryへ委譲して作成パスを返す", func(t *testing.T) {
		readDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		shelf := &recordingShelf{newPath: "Some Book.toml"}
		service := NewService(shelf, fakeQuery{})

		output, err := service.Run(Command{Kind: New, ShelfRoot: "shelf", Title: "Some Book", ReadDate: readDate})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.Message != "Created Some Book.toml." {
			t.Fatalf("message = %q", output.Message)
		}
		if shelf.newRoot != "shelf" || shelf.newTitle != "Some Book" || !shelf.newDate.Equal(readDate) {
			t.Fatalf("new call = %#v", shelf)
		}
	})
}

func validShelfData() ShelfData {
	return ShelfData{
		Config: AllowedValuesData{
			Genres:     []string{"Novel", "Technical"},
			Publishers: []string{"Example Publisher"},
			Imprints:   []string{"Example Paperback"},
			Editions: []EditionData{
				{Name: "Paperback", ImprintRequired: true},
				{Name: "Hardcover", ImprintRequired: false},
			},
		},
		Books: []BookData{
			{
				Path:      "B.toml",
				Title:     "B",
				Author:    "Author B",
				Rating:    90,
				ReadDate:  time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				Genre:     "Novel",
				Publisher: "Example Publisher",
				Edition:   "Paperback",
				Imprint:   "Example Paperback",
			},
			{
				Path:      "A.toml",
				Title:     "A",
				Author:    "Author A",
				Rating:    80,
				ReadDate:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Genre:     "Technical",
				Publisher: "Example Publisher",
			},
		},
	}
}

type fakeShelf struct {
	data ShelfData
	err  error
}

func (shelf fakeShelf) Init(string) error {
	return shelf.err
}

func (shelf fakeShelf) CreateBook(string, string, time.Time) (string, error) {
	return "", shelf.err
}

func (shelf fakeShelf) Load(string) (ShelfData, error) {
	return shelf.data, shelf.err
}

type recordingShelf struct {
	initRoot string
	newRoot  string
	newTitle string
	newDate  time.Time
	newPath  string
}

func (shelf *recordingShelf) Init(root string) error {
	shelf.initRoot = root
	return nil
}

func (shelf *recordingShelf) CreateBook(root, title string, readDate time.Time) (string, error) {
	shelf.newRoot = root
	shelf.newTitle = title
	shelf.newDate = readDate
	if shelf.newPath == "" {
		return "", errors.New("missing new path")
	}
	return shelf.newPath, nil
}

func (shelf *recordingShelf) Load(string) (ShelfData, error) {
	return ShelfData{}, nil
}

type fakeQuery struct{}

func (fakeQuery) List([]domain.Book) (Table, error) {
	return Table{}, nil
}

func (fakeQuery) Query([]domain.Book, string) (Table, error) {
	return Table{}, nil
}

type recordingQuery struct {
	table     Table
	listCount int
	where     string
	books     []domain.Book
}

func (query *recordingQuery) List(books []domain.Book) (Table, error) {
	query.listCount++
	query.books = books
	return query.table, nil
}

func (query *recordingQuery) Query(books []domain.Book, where string) (Table, error) {
	query.books = books
	query.where = where
	return query.table, nil
}
