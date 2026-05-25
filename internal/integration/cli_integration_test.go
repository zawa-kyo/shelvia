package integration_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zawa-kyo/shelvia/internal/adapter/cli"
	"github.com/zawa-kyo/shelvia/internal/adapter/localfs"
	"github.com/zawa-kyo/shelvia/internal/adapter/queryengine"
	"github.com/zawa-kyo/shelvia/internal/application"
)

func TestCLIRunWithRealAdapters(t *testing.T) {
	t.Run("initからvalidate/list/queryまで同じshelfを扱える", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "reading-log")
		service := application.NewService(localfs.Repository{}, queryengine.Store{})

		code, stdout, stderr := runCLI(service, "init", root)

		require.Equal(t, 0, code, "stderr = %q", stderr)
		require.Contains(t, stdout, "Initialized shelf.")

		t.Setenv("SHELVIA_DIR", root)
		code, stdout, stderr = runCLI(service, "validate")

		require.Equal(t, 0, code, "stderr = %q", stderr)
		require.Equal(t, "Validated 1 book, 1 config file.\n", stdout)

		code, stdout, stderr = runCLI(service, "list")

		require.Equal(t, 0, code, "stderr = %q", stderr)
		require.Contains(t, stdout, "Example Book")
		require.Contains(t, stdout, "read_date")

		code, stdout, stderr = runCLI(service, "query", "--where", `rating >= 80`)

		require.Equal(t, 0, code, "stderr = %q", stderr)
		require.Contains(t, stdout, "Example Book")
	})
}

func runCLI(service application.Service, args ...string) (int, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := cli.Run(args, service, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}
