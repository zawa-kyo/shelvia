//go:build e2e

package e2e_test

import (
	"bytes"
	"errors"
	"fmt"
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
	shelf := filepath.Join(t.TempDir(), "shelf")

	initResult := runShelvia(t, bin, nil, "init", shelf)
	require.Equal(t, 0, initResult.code, "stderr = %q", initResult.stderr)
	require.Contains(t, initResult.stdout, "Initialized shelf.")

	validateResult := runShelvia(t, bin, []string{"SHELVIA_DIR=" + shelf}, "validate")
	require.Equal(t, 0, validateResult.code, "stderr = %q", validateResult.stderr)
	require.Contains(t, validateResult.stdout, "Validated 1 book, 1 config file.")

	listResult := runShelvia(t, bin, []string{"SHELVIA_DIR=" + shelf}, "list")
	require.Equal(t, 0, listResult.code, "stderr = %q", listResult.stderr)
	require.Contains(t, listResult.stdout, "read_date")
	require.Contains(t, listResult.stdout, "Example Book")

	queryResult := runShelvia(t, bin, []string{"SHELVIA_DIR=" + shelf}, "query", "--where", "rating >= 80")
	require.Equal(t, 0, queryResult.code, "stderr = %q", queryResult.stderr)
	require.Contains(t, queryResult.stdout, "Example Book")
}

func TestShelviaCLISortsFixtureShelf(t *testing.T) {
	repoRoot := repoRoot(t)
	bin := buildShelvia(t, repoRoot)
	fixtureShelf := filepath.Join(repoRoot, "fixtures", "shelves", "valid", "minimal")

	validateResult := runShelvia(t, bin, []string{"SHELVIA_DIR=" + fixtureShelf}, "validate")
	require.Equal(t, 0, validateResult.code, "stderr = %q", validateResult.stderr)
	bookCount := countBookTOMLFiles(t, fixtureShelf)
	require.Contains(t, validateResult.stdout, fmt.Sprintf("Validated %d book%s, 1 config file.", bookCount, pluralSuffix(bookCount)))

	sortResult := runShelvia(t, bin, []string{"SHELVIA_DIR=" + fixtureShelf}, "query", "--where", "rating >= 0 order by rating desc")
	require.Equal(t, 0, sortResult.code, "stderr = %q", sortResult.stderr)
	requireContainsInOrder(t, sortResult.stdout, "Top Rated Book", "Some Book", "Example Book", "Mid Rated Book", "Low Rated Book")
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

func runShelvia(t *testing.T, bin string, env []string, args ...string) commandResult {
	t.Helper()

	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), env...)
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

func requireContainsInOrder(t *testing.T, output string, values ...string) {
	t.Helper()

	offset := 0
	for _, value := range values {
		index := strings.Index(output[offset:], value)
		require.NotEqualf(t, -1, index, "expected %q after byte offset %d in output:\n%s", value, offset, output)
		offset += index + len(value)
	}
}

func countBookTOMLFiles(t *testing.T, root string) int {
	t.Helper()

	entries, err := os.ReadDir(root)
	require.NoError(t, err)

	count := 0
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "config.toml" || filepath.Ext(entry.Name()) != ".toml" {
			continue
		}
		count++
	}
	return count
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	return filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
}
