package application

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zawa-kyo/shelvia/internal/domain"
)

func TestServiceRunValidate(t *testing.T) {
	t.Run("shelfを読み込んで検証件数を返す", func(t *testing.T) {
		service := NewService(fakeShelf{data: validShelfData()}, fakeQuery{})
		command := Command{Kind: Validate, ShelfRoot: "shelf"}

		output, err := service.Run(command)

		require.NoError(t, err)
		require.Equal(t, "Validated 2 books, 1 config file.", output.Message)
	})

	t.Run("domain validation errorにはファイルパスを付ける", func(t *testing.T) {
		data := validShelfData()
		data.Books[0].Genre = "Unknown"
		service := NewService(fakeShelf{data: data}, fakeQuery{})
		command := Command{Kind: Validate, ShelfRoot: "shelf"}

		_, err := service.Run(command)

		require.EqualError(t, err, "B.toml: genre: unknown genre: Unknown")
	})
}

func TestServiceRunListAndSearch(t *testing.T) {
	t.Run("listは検証済みbookをquery portへ渡す", func(t *testing.T) {
		query := &recordingQuery{table: Table{Headers: []string{"title"}, Rows: [][]string{{"A"}}}}
		service := NewService(fakeShelf{data: validShelfData()}, query)
		command := Command{Kind: List, ShelfRoot: "shelf"}

		output, err := service.Run(command)

		require.NoError(t, err)
		require.Equal(t, query.table, output.Table)
		require.Equal(t, 1, query.listCount)
		require.Len(t, query.books, 2)
	})

	t.Run("searchはwhere fragmentをquery portへ渡す", func(t *testing.T) {
		query := &recordingQuery{table: Table{Headers: []string{"title"}, Rows: [][]string{{"B"}}}}
		service := NewService(fakeShelf{data: validShelfData()}, query)
		command := Command{Kind: Search, ShelfRoot: "shelf", Where: `rating >= 90`}

		_, err := service.Run(command)

		require.NoError(t, err)
		require.Equal(t, `rating >= 90`, query.where)
	})

	t.Run("shelfの読み込みに失敗した場合はquery portを呼ばない", func(t *testing.T) {
		query := &recordingQuery{}
		service := NewService(fakeShelf{err: errors.New("load failed")}, query)
		command := Command{Kind: List, ShelfRoot: "shelf"}

		_, err := service.Run(command)

		require.EqualError(t, err, "load failed")
		require.Zero(t, query.listCount)
		require.Empty(t, query.books)
	})

	t.Run("query portのエラーを返す", func(t *testing.T) {
		query := &recordingQuery{err: errors.New("query failed")}
		service := NewService(fakeShelf{data: validShelfData()}, query)
		command := Command{Kind: Search, ShelfRoot: "shelf", Where: `rating >= 90`}

		_, err := service.Run(command)

		require.EqualError(t, err, "query failed")
		require.Equal(t, `rating >= 90`, query.where)
		require.Len(t, query.books, 2)
	})
}

func TestServiceRunInitAndAdd(t *testing.T) {
	t.Run("initはrepositoryへ委譲する", func(t *testing.T) {
		shelf := &recordingShelf{newPath: "Some Book.toml"}
		service := NewService(shelf, fakeQuery{})
		command := Command{Kind: Init, ShelfRoot: "shelf"}

		_, err := service.Run(command)

		require.NoError(t, err)
		require.Equal(t, "shelf", shelf.initRoot)
	})

	t.Run("addはrepositoryへ委譲して作成パスを返す", func(t *testing.T) {
		readDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		shelf := &recordingShelf{newPath: "Some Book.toml"}
		service := NewService(shelf, fakeQuery{})
		command := Command{Kind: Add, ShelfRoot: "shelf", Title: "Some Book", ReadDate: readDate}

		output, err := service.Run(command)

		require.NoError(t, err)
		require.Equal(t, "Created Some Book.toml.", output.Message)
		require.Equal(t, "shelf", shelf.newRoot)
		require.Equal(t, "Some Book", shelf.newTitle)
		require.True(t, shelf.newDate.Equal(readDate))
	})
}

func TestServiceRunShowAndPath(t *testing.T) {
	t.Run("showはtitleに一致する1冊の詳細を返す", func(t *testing.T) {
		service := NewService(fakeShelf{data: validShelfData()}, fakeQuery{})
		command := Command{Kind: Show, ShelfRoot: "shelf", Title: "A"}

		output, err := service.Run(command)

		require.NoError(t, err)
		require.Equal(t, []string{"field", "value"}, output.Table.Headers)
		require.Contains(t, output.Table.Rows, []string{"title", "A"})
		require.Contains(t, output.Table.Rows, []string{"path", "A.toml"})
	})

	t.Run("pathはtitleに一致する1冊のファイルパスを返す", func(t *testing.T) {
		service := NewService(fakeShelf{data: validShelfData()}, fakeQuery{})
		command := Command{Kind: Path, ShelfRoot: "shelf", Title: "A"}

		output, err := service.Run(command)

		require.NoError(t, err)
		require.Equal(t, "A.toml", output.Message)
	})

	t.Run("一致するtitleがなければエラーにする", func(t *testing.T) {
		service := NewService(fakeShelf{data: validShelfData()}, fakeQuery{})
		command := Command{Kind: Show, ShelfRoot: "shelf", Title: "Missing"}

		_, err := service.Run(command)

		require.EqualError(t, err, `book title not found: Missing`)
	})

	t.Run("同じtitleが複数あれば候補パスを含むエラーにする", func(t *testing.T) {
		data := validShelfData()
		data.Books[1].Title = "B"
		service := NewService(fakeShelf{data: data}, fakeQuery{})
		command := Command{Kind: Path, ShelfRoot: "shelf", Title: "B"}

		_, err := service.Run(command)

		require.EqualError(t, err, `book title is ambiguous: B (B.toml, A.toml)`)
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
	err       error
	listCount int
	where     string
	books     []domain.Book
}

func (query *recordingQuery) List(books []domain.Book) (Table, error) {
	query.listCount++
	query.books = books
	return query.table, query.err
}

func (query *recordingQuery) Query(books []domain.Book, where string) (Table, error) {
	query.books = books
	query.where = where
	return query.table, query.err
}
