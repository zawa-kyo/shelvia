package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zawa-kyo/shelvia/internal/application"
)

func TestParse(t *testing.T) {
	t.Run("明示的なshelfが環境変数より優先される", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")
		args := []string{"--shelf", "arg-shelf", "validate"}

		command, err := Parse(args)

		require.NoError(t, err)
		require.Equal(t, application.Validate, command.Kind)
		require.Equal(t, "arg-shelf", command.ShelfRoot)
	})

	t.Run("shelf未指定なら環境変数を参照する", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")
		args := []string{"list"}

		command, err := Parse(args)

		require.NoError(t, err)
		require.Equal(t, application.List, command.Kind)
		require.Equal(t, "env-shelf", command.ShelfRoot)
	})

	t.Run("どちらもない場合はエラーにする", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "")
		require.NoError(t, os.Unsetenv("SHELVIA_DIR"))
		args := []string{"validate"}

		_, err := Parse(args)

		require.Error(t, err)
	})

	t.Run("initはPATHをshelf rootとして使う", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")
		args := []string{"init", "new-shelf"}

		command, err := Parse(args)

		require.NoError(t, err)
		require.Equal(t, application.Init, command.Kind)
		require.Equal(t, "new-shelf", command.ShelfRoot)
	})

	t.Run("initはPATH省略時にカレントディレクトリを使う", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")
		args := []string{"init"}

		command, err := Parse(args)

		require.NoError(t, err)
		require.Equal(t, application.Init, command.Kind)
		require.Equal(t, ".", command.ShelfRoot)
	})

	t.Run("addはタイトルと日付を読む", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")
		args := []string{"add", "--date", "2024-01-02", "Some Book"}

		command, err := Parse(args)

		require.NoError(t, err)
		require.Equal(t, application.Add, command.Kind)
		require.Equal(t, "Some Book", command.Title)
		require.Equal(t, "2024-01-02", command.ReadDate.Format("2006-01-02"))
	})

	t.Run("searchはwhere fragmentを読む", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")
		args := []string{"search", "rating >= 90"}

		command, err := Parse(args)

		require.NoError(t, err)
		require.Equal(t, application.Search, command.Kind)
		require.Equal(t, "rating >= 90", command.Where)
	})

	t.Run("showはタイトルを読む", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")
		args := []string{"show", "Some Book"}

		command, err := Parse(args)

		require.NoError(t, err)
		require.Equal(t, application.Show, command.Kind)
		require.Equal(t, "Some Book", command.Title)
	})

	t.Run("pathはタイトルを読む", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")
		args := []string{"path", "Some Book"}

		command, err := Parse(args)

		require.NoError(t, err)
		require.Equal(t, application.Path, command.Kind)
		require.Equal(t, "Some Book", command.Title)
	})

	t.Run("newはサポートしない", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")
		args := []string{"new", "Some Book", "--read-date", "2024-01-02"}

		_, err := Parse(args)

		require.Error(t, err)
	})

	t.Run("queryはサポートしない", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")
		args := []string{"query", "--where", "rating >= 90"}

		_, err := Parse(args)

		require.Error(t, err)
	})
}
