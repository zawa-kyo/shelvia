package integration_test

import (
	"bytes"
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
