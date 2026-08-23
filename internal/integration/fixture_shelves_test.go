package integration_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zawa-kyo/shelvia/internal/adapter/localfs"
	"github.com/zawa-kyo/shelvia/internal/adapter/queryengine"
	"github.com/zawa-kyo/shelvia/internal/application"
)

func TestFixtureShelves(t *testing.T) {
	t.Run("評価が0点の書籍を含む正常な最小構成のshelfを検証できる", func(t *testing.T) {
		service := application.NewService(localfs.Repository{}, queryengine.Store{})
		command := application.Command{
			Kind:      application.Validate,
			ShelfRoot: fixturePath(t, "valid", "minimal"),
		}

		output, err := service.Run(command)

		require.NoError(t, err)
		require.Equal(t, "Validated 5 books, 1 config file.", output.Message)
	})

	tests := []struct {
		name string
		want string
	}{
		{name: "nested-config", want: "nested/config.toml: misplaced config.toml"},
		{name: "missing-required", want: "book.toml: title"},
		{name: "missing-rating", want: "book.toml: rating: required"},
		{name: "unknown-config-value", want: "book.toml: genre: unknown genre"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"は検証エラーになる", func(t *testing.T) {
			service := application.NewService(localfs.Repository{}, queryengine.Store{})
			command := application.Command{
				Kind:      application.Validate,
				ShelfRoot: fixturePath(t, "invalid", tt.name),
			}

			_, err := service.Run(command)

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.want)
		})
	}
}

func fixturePath(t *testing.T, parts ...string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")

	root := filepath.Join(filepath.Dir(file), "..", "..", "fixtures", "shelves")
	all := append([]string{root}, parts...)
	return filepath.Clean(filepath.Join(all...))
}
