package localfs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRepositoryInitAndLoad(t *testing.T) {
	t.Run("存在しないディレクトリを作成して初期化する", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "new-shelf")

		err := Repository{}.Init(root)

		require.NoError(t, err)
		require.FileExists(t, filepath.Join(root, "config.toml"))
		require.FileExists(t, filepath.Join(root, "example.toml"))
	})

	t.Run("init直後のshelfは検証可能なデータとして読み込める", func(t *testing.T) {
		root := t.TempDir()
		repository := Repository{}

		err := repository.Init(root)
		require.NoError(t, err)

		data, err := repository.Load(root)

		require.NoError(t, err)
		require.Equal(t, []string{"Novel"}, data.Config.Genres)
		require.Len(t, data.Books, 1)
		require.Equal(t, "example.toml", data.Books[0].Path)
	})

	t.Run("initが生成するconfig.tomlには編集の助けになるコメントを含む", func(t *testing.T) {
		root := t.TempDir()

		err := Repository{}.Init(root)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(root, "config.toml"))
		require.NoError(t, err)
		require.Contains(t, string(content), "# Add the genres you use.")
		require.Contains(t, string(content), "# Add imprint names if you use them. Leave this as [] if not needed yet.")
	})

	for _, filename := range []string{"config.toml", "example.toml"} {
		t.Run(filename+"が存在する場合は何も上書きも作成もしない", func(t *testing.T) {
			root := t.TempDir()
			existingPath := filepath.Join(root, filename)
			require.NoError(t, os.WriteFile(existingPath, []byte("already here"), 0o644))

			err := Repository{}.Init(root)

			require.ErrorContains(t, err, "already exists")
			content, readErr := os.ReadFile(existingPath)
			require.NoError(t, readErr)
			require.Equal(t, "already here", string(content))
			entries, readDirErr := os.ReadDir(root)
			require.NoError(t, readDirErr)
			require.Len(t, entries, 1)
		})
	}
}

func TestRepositoryCreateBook(t *testing.T) {
	t.Run("安全なファイル名でTOMLテンプレートを作成する", func(t *testing.T) {
		root := t.TempDir()
		readDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

		path, err := Repository{}.CreateBook(root, "Some Book", readDate)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(root, path))

		require.NoError(t, err)
		require.Equal(t, "Some Book.toml", path)
		require.Contains(t, string(content), `title = "Some Book"`)
		require.Contains(t, string(content), "read_date = 2024-01-02")
	})

	t.Run("toml拡張子は二重に付けない", func(t *testing.T) {
		root := t.TempDir()

		path, err := Repository{}.CreateBook(root, "Some Book.toml", time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC))

		require.NoError(t, err)
		require.Equal(t, "Some Book.toml", path)
		require.FileExists(t, filepath.Join(root, path))
	})

	t.Run("同名ファイルは上書きしない", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, "Some Book.toml")
		require.NoError(t, os.WriteFile(path, []byte("already here"), 0o644))

		_, err := Repository{}.CreateBook(root, "Some Book", time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC))

		require.ErrorContains(t, err, "already exists")
		content, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		require.Equal(t, "already here", string(content))
	})

	t.Run("Windowsでも危険なファイル名を拒否する", func(t *testing.T) {
		root := t.TempDir()
		invalidTitles := []struct {
			name  string
			title string
		}{
			{name: "空のタイトル", title: ""},
			{name: "現在ディレクトリを表す名前", title: "."},
			{name: "親ディレクトリを表す名前", title: ".."},
			{name: "Windowsの予約名", title: "CON"},
			{name: "拡張子付きのWindows予約名", title: "COM1.txt"},
			{name: "スラッシュを含む名前", title: "a/b"},
			{name: "バックスラッシュを含む名前", title: `a\b`},
			{name: "コロンを含む名前", title: "a:b"},
			{name: "アスタリスクを含む名前", title: "a*"},
			{name: "末尾がドットの名前", title: "name."},
		}
		for _, tt := range invalidTitles {
			t.Run(tt.name, func(t *testing.T) {
				_, err := Repository{}.CreateBook(root, tt.title, time.Now())

				require.Error(t, err)
			})
		}
	})
}

func TestRepositoryLoad(t *testing.T) {
	t.Run("直下以外のconfig.tomlは誤配置として拒否する", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, (Repository{}).Init(root))

		nested := filepath.Join(root, "nested")
		require.NoError(t, os.Mkdir(nested, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(nested, "config.toml"), []byte(`kind = "shelvia-config"`), 0o644))

		_, err := Repository{}.Load(root)

		require.Error(t, err)
		require.Contains(t, err.Error(), "nested/config.toml: misplaced config.toml")
	})

	t.Run("TOMLを再帰的に読み込む", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, (Repository{}).Init(root))

		subdir := filepath.Join(root, "2024")
		require.NoError(t, os.Mkdir(subdir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(subdir, "Nested.toml"), []byte(validBookTOML("Nested")), 0o644))

		data, err := Repository{}.Load(root)

		require.NoError(t, err)

		paths := map[string]struct{}{}
		for _, book := range data.Books {
			paths[book.Path] = struct{}{}
		}

		require.Len(t, data.Books, 2)
		require.Contains(t, paths, "2024/Nested.toml")
		require.Contains(t, paths, "example.toml")
	})

	t.Run("TOML parse errorはshelf rootからの相対パスで表示する", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, (Repository{}).Init(root))

		subdir := filepath.Join(root, "2024")
		require.NoError(t, os.Mkdir(subdir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(subdir, "Broken.toml"), []byte("title = "), 0o644))

		_, err := Repository{}.Load(root)

		require.Error(t, err)
		require.NotContains(t, err.Error(), root)
		require.Contains(t, err.Error(), "2024/Broken.toml")
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
