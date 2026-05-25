package localfs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRepositoryInitAndLoad(t *testing.T) {
	t.Run("init直後のshelfは検証可能なデータとして読み込める", func(t *testing.T) {
		root := t.TempDir()
		repository := Repository{}

		if err := repository.Init(root); err != nil {
			t.Fatalf("Init: %v", err)
		}
		data, err := repository.Load(root)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if len(data.Config.Genres) != 1 || data.Config.Genres[0] != "Novel" {
			t.Fatalf("genres = %#v", data.Config.Genres)
		}
		if len(data.Books) != 1 {
			t.Fatalf("books = %d, want 1", len(data.Books))
		}
		if data.Books[0].Path != "example.toml" {
			t.Fatalf("book path = %q", data.Books[0].Path)
		}
	})

	t.Run("既存ファイルは上書きしない", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "config.toml"), []byte("already here"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		err := Repository{}.Init(root)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "already exists") {
			t.Fatalf("error = %q", err.Error())
		}
	})
}

func TestRepositoryCreateBook(t *testing.T) {
	t.Run("安全なファイル名でTOMLテンプレートを作成する", func(t *testing.T) {
		root := t.TempDir()
		readDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

		path, err := Repository{}.CreateBook(root, "Some Book", readDate)
		if err != nil {
			t.Fatalf("CreateBook: %v", err)
		}
		if path != "Some Book.toml" {
			t.Fatalf("path = %q", path)
		}
		content, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if !strings.Contains(string(content), `title = "Some Book"`) {
			t.Fatalf("content does not contain title: %s", content)
		}
		if !strings.Contains(string(content), "read_date = 2024-01-02") {
			t.Fatalf("content does not contain read_date: %s", content)
		}
	})

	t.Run("Windowsでも危険なファイル名を拒否する", func(t *testing.T) {
		root := t.TempDir()
		invalidTitles := []string{"", ".", "..", "CON", "COM1.txt", "a/b", `a\b`, "a:b", "a*", "name."}
		for _, title := range invalidTitles {
			t.Run(title, func(t *testing.T) {
				_, err := Repository{}.CreateBook(root, title, time.Now())
				if err == nil {
					t.Fatal("expected error")
				}
			})
		}
	})
}

func TestRepositoryLoad(t *testing.T) {
	t.Run("直下以外のconfig.tomlは誤配置として拒否する", func(t *testing.T) {
		root := t.TempDir()
		if err := (Repository{}).Init(root); err != nil {
			t.Fatalf("Init: %v", err)
		}
		nested := filepath.Join(root, "nested")
		if err := os.Mkdir(nested, 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(nested, "config.toml"), []byte(`kind = "shelvia-config"`), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		_, err := Repository{}.Load(root)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "nested/config.toml: misplaced config.toml") {
			t.Fatalf("error = %q", err.Error())
		}
	})

	t.Run("TOMLを再帰的に読み込む", func(t *testing.T) {
		root := t.TempDir()
		if err := (Repository{}).Init(root); err != nil {
			t.Fatalf("Init: %v", err)
		}
		subdir := filepath.Join(root, "2024")
		if err := os.Mkdir(subdir, 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(subdir, "Nested.toml"), []byte(validBookTOML("Nested")), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		data, err := Repository{}.Load(root)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if len(data.Books) != 2 {
			t.Fatalf("books = %d, want 2", len(data.Books))
		}
		paths := map[string]struct{}{}
		for _, book := range data.Books {
			paths[book.Path] = struct{}{}
		}
		for _, want := range []string{"2024/Nested.toml", "example.toml"} {
			if _, ok := paths[want]; !ok {
				t.Fatalf("book paths = %#v, want %s", paths, want)
			}
		}
	})

	t.Run("TOML parse errorはshelf rootからの相対パスで表示する", func(t *testing.T) {
		root := t.TempDir()
		if err := (Repository{}).Init(root); err != nil {
			t.Fatalf("Init: %v", err)
		}
		subdir := filepath.Join(root, "2024")
		if err := os.Mkdir(subdir, 0o755); err != nil {
			t.Fatalf("Mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(subdir, "Broken.toml"), []byte("title = "), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		_, err := Repository{}.Load(root)
		if err == nil {
			t.Fatal("expected error")
		}
		if strings.Contains(err.Error(), root) {
			t.Fatalf("error should not contain absolute root: %q", err.Error())
		}
		if !strings.Contains(err.Error(), "2024/Broken.toml") {
			t.Fatalf("error = %q, want relative path", err.Error())
		}
	})
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
