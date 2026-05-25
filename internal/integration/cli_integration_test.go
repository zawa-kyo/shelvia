package integration_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

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
		if code != 0 {
			t.Fatalf("init code = %d, stderr = %q", code, stderr)
		}
		if !strings.Contains(stdout, "Initialized shelf.") {
			t.Fatalf("init stdout = %q", stdout)
		}

		t.Setenv("SHELVIA_DIR", root)
		code, stdout, stderr = runCLI(service, "validate")
		if code != 0 {
			t.Fatalf("validate code = %d, stderr = %q", code, stderr)
		}
		if stdout != "Validated 1 book, 1 config file.\n" {
			t.Fatalf("validate stdout = %q", stdout)
		}

		code, stdout, stderr = runCLI(service, "list")
		if code != 0 {
			t.Fatalf("list code = %d, stderr = %q", code, stderr)
		}
		if !strings.Contains(stdout, "Example Book") || !strings.Contains(stdout, "read_date") {
			t.Fatalf("list stdout = %q", stdout)
		}

		code, stdout, stderr = runCLI(service, "query", "--where", `rating >= 80`)
		if code != 0 {
			t.Fatalf("query code = %d, stderr = %q", code, stderr)
		}
		if !strings.Contains(stdout, "Example Book") {
			t.Fatalf("query stdout = %q", stdout)
		}
	})
}

func runCLI(service application.Service, args ...string) (int, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := cli.Run(args, service, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}
