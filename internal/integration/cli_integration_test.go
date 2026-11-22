package integration_test

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zawa-kyo/shelvia/internal/adapter/cli"
	"github.com/zawa-kyo/shelvia/internal/adapter/localfs"
	"github.com/zawa-kyo/shelvia/internal/adapter/queryengine"
	"github.com/zawa-kyo/shelvia/internal/application"
)

func TestCLIRunWithRealAdapters(t *testing.T) {
	t.Run("initからadd/validate/list/search/show/pathまで同じshelfを扱える", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "reading-log")
		service := application.NewService(localfs.Repository{}, queryengine.Store{})

		code, stdout, stderr := runCLI(service, "init", root)

		require.Equal(t, 0, code, "stderr = %q", stderr)
		require.Contains(t, stdout, "Initialized shelf.")

		t.Setenv("SHELVIA_DIR", root)
		code, stdout, stderr = runCLI(service, "add", "Some Book", "--date", "2024-01-02")

		require.Equal(t, 0, code, "stderr = %q", stderr)
		require.Contains(t, stdout, "Created Some Book.toml.")
		require.NoError(t, os.WriteFile(filepath.Join(root, "Some Book.toml"), []byte(validBookTOML("Some Book")), 0o644))

		code, stdout, stderr = runCLI(service, "validate")

		require.Equal(t, 0, code, "stderr = %q", stderr)
		require.Equal(t, "Validated 2 books, 1 config file.\n", stdout)

		code, stdout, stderr = runCLI(service, "list")

		require.Equal(t, 0, code, "stderr = %q", stderr)
		require.Contains(t, stdout, "Example Book")
		require.Contains(t, stdout, "Some Book")
		require.Contains(t, stdout, "read_date")
		require.Less(t, strings.Index(stdout, "Some Book"), strings.Index(stdout, "Example Book"))

		code, stdout, stderr = runCLI(service, "search", `title = "Some Book"`)

		require.Equal(t, 0, code, "stderr = %q", stderr)
		require.Contains(t, stdout, "Some Book")
		require.NotContains(t, stdout, "Example Book")

		code, stdout, stderr = runCLI(service, "show", "Some Book")

		require.Equal(t, 0, code, "stderr = %q", stderr)
		require.Contains(t, stdout, "title")
		require.Contains(t, stdout, "Some Book")

		code, stdout, stderr = runCLI(service, "path", "Some Book")

		require.Equal(t, 0, code, "stderr = %q", stderr)
		require.Equal(t, "Some Book.toml\n", stdout)
	})

	t.Run("adapterから返されたファイルパス付き診断をstderrへ表示する", func(t *testing.T) {
		root := t.TempDir()
		service := application.NewService(localfs.Repository{}, queryengine.Store{})
		require.NoError(t, localfs.Repository{}.Init(root))
		require.NoError(t, os.WriteFile(filepath.Join(root, "Broken.toml"), []byte("title = "), 0o644))

		code, stdout, stderr := runCLI(service, "--shelf", root, "validate")

		require.Equal(t, 1, code)
		require.Empty(t, stdout)
		require.Contains(t, stderr, "Broken.toml")
		require.NotContains(t, stderr, root)
	})
}

func TestCSVOutput(t *testing.T) {
	root := t.TempDir()
	service := application.NewService(localfs.Repository{}, queryengine.Store{})
	require.NoError(t, localfs.Repository{}.Init(root))
	book := validBookTOML("A very long book title, with comma") + `
[thoughts]
body = """First line
Second line"""
`
	require.NoError(t, os.WriteFile(filepath.Join(root, "long.toml"), []byte(book), 0o644))
	t.Setenv("SHELVIA_DIR", root)

	commands := [][]string{
		{"list", "--format", "csv"},
		{"search", `title like "A very%"`, "--format", "csv"},
		{"export"},
	}
	for _, args := range commands {
		code, stdout, stderr := runCLI(service, args...)
		require.Equal(t, 0, code, stderr)
		rows, err := csv.NewReader(strings.NewReader(stdout)).ReadAll()
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(rows), 2)
		require.Contains(t, stdout, "A very long book title, with comma")
		if args[0] == "export" {
			require.Equal(t, "thoughts.body", rows[0][11])
			require.Contains(t, stdout, "First line\nSecond line")
		} else {
			require.Len(t, rows[0], 6)
		}
	}
}

func TestMultipleValues(t *testing.T) {
	root := t.TempDir()
	service := application.NewService(localfs.Repository{}, queryengine.Store{})
	require.NoError(t, localfs.Repository{}.Init(root))
	data := `title = "Collaboration"
author = ["Alice, A", "Bob"]
genre = ["Novel"]
translator = ["Taro", "Jiro"]
rating = 90
read_date = 2024-01-02
publisher = "Example Publisher"
`
	require.NoError(t, os.WriteFile(filepath.Join(root, "book.toml"), []byte(data), 0o644))
	t.Setenv("SHELVIA_DIR", root)
	for _, condition := range []string{`author = "Bob"`, `translator = "Jiro"`, `genre = "Novel"`, `author != "Alice, A"`} {
		code, stdout, stderr := runCLI(service, "search", condition, "--format", "csv")
		require.Equal(t, 0, code, stderr)
		if condition == `author != "Alice, A"` {
			require.NotContains(t, stdout, "Collaboration")
		} else {
			require.Contains(t, stdout, "Collaboration")
		}
	}
	code, stdout, stderr := runCLI(service, "export")
	require.Equal(t, 0, code, stderr)
	rows, err := csv.NewReader(strings.NewReader(stdout)).ReadAll()
	require.NoError(t, err)
	for _, row := range rows {
		if row[0] == "Collaboration" {
			require.Equal(t, `["Alice, A","Bob"]`, row[1])
			require.Equal(t, `["Taro","Jiro"]`, row[9])
			return
		}
	}
	t.Fatal("export did not include Collaboration")
}

func TestShowDisplaysCompleteDetails(t *testing.T) {
	root := t.TempDir()
	service := application.NewService(localfs.Repository{}, queryengine.Store{})
	require.NoError(t, localfs.Repository{}.Init(root))
	data := `title = "Long Details"
author = ["A very long author name", "Another very long author name"]
genre = ["Novel"]
translator = ["A very long translator name", "Another long translator name"]
rating = 90
read_date = 2024-01-02
publisher = "Example Publisher"
`
	require.NoError(t, os.WriteFile(filepath.Join(root, "book.toml"), []byte(data), 0o644))
	t.Setenv("SHELVIA_DIR", root)
	code, stdout, stderr := runCLI(service, "show", "Long Details")
	require.Equal(t, 0, code, stderr)
	require.Contains(t, stdout, "Another very long author name")
	require.Contains(t, stdout, "Another long translator name")
}

func validBookTOML(title string) string {
	return `title = "` + title + `"
author = "Some Author"
rating = 90
read_date = 2024-01-02

genre = "Novel"
edition = "Paperback"
imprint = "Example Paperback"
publisher = "Example Publisher"
`
}

func runCLI(service application.Service, args ...string) (int, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := cli.Run(args, service, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestBookOutputFormats(t *testing.T) {
	root := t.TempDir()
	service := application.NewService(localfs.Repository{}, queryengine.Store{})
	require.NoError(t, localfs.Repository{}.Init(root))
	require.NoError(t, os.Remove(filepath.Join(root, "example.toml")))
	data := `title = "A, book"
author = ["Alice", "Bob"]
genre = ["Novel"]
translator = ["Taro", "Jiro"]
rating = 90
read_date = 2024-01-02
publisher = "Example Publisher"

[thoughts]
summary = "A summary"
body = "A body"
`
	require.NoError(t, os.WriteFile(filepath.Join(root, "book.toml"), []byte(data), 0o644))
	t.Setenv("SHELVIA_DIR", root)

	commands := [][]string{
		{"list", "--format", "csv"},
		{"search", `title = "A, book"`, "--format", "csv"},
	}
	for _, args := range commands {
		code, stdout, stderr := runCLI(service, args...)
		require.Equal(t, 0, code, stderr)
		rows, err := csv.NewReader(strings.NewReader(stdout)).ReadAll()
		require.NoError(t, err)
		require.Equal(t, [][]string{
			{"read_date", "rating", "title", "author", "genre", "publisher"},
			{"2024-01-02", "90", "A, book", `["Alice","Bob"]`, `["Novel"]`, "Example Publisher"},
		}, rows)
	}

	code, stdout, stderr := runCLI(service, "export")
	require.Equal(t, 0, code, stderr)
	rows, err := csv.NewReader(strings.NewReader(stdout)).ReadAll()
	require.NoError(t, err)
	require.Equal(t, [][]string{
		{
			"title", "author", "rating", "read_date", "genre", "publisher",
			"edition", "imprint", "series", "translator", "thoughts.summary", "thoughts.body",
		},
		{
			"A, book", `["Alice","Bob"]`, "90", "2024-01-02", `["Novel"]`, "Example Publisher",
			"", "", "", `["Taro","Jiro"]`, "A summary", "A body",
		},
	}, rows)

	code, stdout, stderr = runCLI(service, "search", `title = "missing"`, "--format", "csv")
	require.Equal(t, 0, code, stderr)
	require.Equal(t, "read_date,rating,title,author,genre,publisher\n", stdout)

	code, stdout, stderr = runCLI(service, "show", "A, book")
	require.Equal(t, 0, code, stderr)
	for _, value := range []string{"book.toml", "A, book", `["Alice","Bob"]`, `["Novel"]`, `["Taro","Jiro"]`, "A summary", "A body"} {
		require.Contains(t, stdout, value)
	}
}
