//go:build e2e

package e2e_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShelviaCLI(t *testing.T) {
	repoRoot := repoRoot(t)
	bin := buildShelvia(t, repoRoot)

	t.Run("initしたshelfでaddからsearchまで実行できる", func(t *testing.T) {
		shelf := filepath.Join(t.TempDir(), "reading-log")

		result := runShelvia(t, bin, "", "init", shelf)

		requireSuccess(t, result, "Initialized shelf.\n")
		require.FileExists(t, filepath.Join(shelf, "config.toml"))
		require.FileExists(t, filepath.Join(shelf, "example.toml"))

		result = runShelvia(t, bin, shelf, "add", "Some Book", "--date", "2024-01-02")

		requireSuccess(t, result, "Created Some Book.toml.\n")
		bookPath := filepath.Join(shelf, "Some Book.toml")
		require.FileExists(t, bookPath)
		require.NoError(t, os.WriteFile(bookPath, []byte(validBookTOML("Some Book")), 0o644))

		missingShelf := filepath.Join(t.TempDir(), "missing")
		result = runShelvia(t, bin, missingShelf, "--shelf", shelf, "validate")

		requireSuccess(t, result, "Validated 2 books, 1 config file.\n")

		result = runShelvia(t, bin, shelf, "list")

		require.Equal(t, 0, result.code, "stderr = %q", result.stderr)
		require.Empty(t, result.stderr)
		require.Contains(t, result.stdout, "read_date")
		requireContainsInOrder(t, result.stdout, "Some Book", "Example Book")

		result = runShelvia(t, bin, shelf, "search", "rating >= 80")

		require.Equal(t, 0, result.code, "stderr = %q", result.stderr)
		require.Empty(t, result.stderr)
		requireContainsInOrder(t, result.stdout, "Some Book", "Example Book")
	})

	t.Run("shelf rootが未指定なら利用方法をstderrに表示する", func(t *testing.T) {
		result := runShelvia(t, bin, "", "validate")

		require.Equal(t, 1, result.code)
		require.Empty(t, result.stdout)
		require.Equal(t, "shelf root is not set; pass --shelf or set SHELVIA_DIR\n", result.stderr)
	})

	t.Run("不正なshelfなら対象ファイルをstderrに表示する", func(t *testing.T) {
		invalidShelf := filepath.Join(repoRoot, "fixtures", "shelves", "invalid", "missing-required")

		result := runShelvia(t, bin, invalidShelf, "validate")

		require.Equal(t, 1, result.code)
		require.Empty(t, result.stdout)
		require.Contains(t, result.stderr, "book.toml")
		require.Contains(t, result.stderr, "title")
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

type commandResult struct {
	code   int
	stdout string
	stderr string
}

func buildShelvia(t *testing.T, repoRoot string) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "shelvia")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/shelvia")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "go build failed:\n%s", string(output))
	return bin
}

func runShelvia(t *testing.T, bin string, shelf string, args ...string) commandResult {
	t.Helper()

	cmd := exec.Command(bin, args...)
	cmd.Env = withoutEnvironmentVariable(os.Environ(), "SHELVIA_DIR")
	if shelf != "" {
		cmd.Env = append(cmd.Env, "SHELVIA_DIR="+shelf)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			code = exitErr.ExitCode()
		} else {
			require.NoError(t, err)
		}
	}
	return commandResult{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func requireSuccess(t *testing.T, result commandResult, stdout string) {
	t.Helper()
	require.Equal(t, 0, result.code, "stderr = %q", result.stderr)
	require.Equal(t, stdout, result.stdout)
	require.Empty(t, result.stderr)
}

func requireContainsInOrder(t *testing.T, output string, values ...string) {
	t.Helper()

	offset := 0
	for _, value := range values {
		index := strings.Index(output[offset:], value)
		require.NotEqualf(t, -1, index, "expected %q after byte offset %d in output:\n%s", value, offset, output)
		offset += index + len(value)
	}
}

func withoutEnvironmentVariable(env []string, name string) []string {
	prefix := name + "="
	filtered := make([]string, 0, len(env))
	for _, value := range env {
		if !strings.HasPrefix(value, prefix) {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	return filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
}
