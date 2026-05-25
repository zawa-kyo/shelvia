package cli

import (
	"os"
	"testing"

	"github.com/zawa-kyo/shelvia/internal/application"
)

func TestParse(t *testing.T) {
	t.Run("明示的なshelfが環境変数より優先される", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")

		command, err := Parse([]string{"--shelf", "arg-shelf", "validate"})
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if command.Kind != application.Validate || command.ShelfRoot != "arg-shelf" {
			t.Fatalf("command = %#v", command)
		}
	})

	t.Run("shelf未指定なら環境変数を参照する", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")

		command, err := Parse([]string{"list"})
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if command.Kind != application.List || command.ShelfRoot != "env-shelf" {
			t.Fatalf("command = %#v", command)
		}
	})

	t.Run("どちらもない場合はエラーにする", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "")
		if err := os.Unsetenv("SHELVIA_DIR"); err != nil {
			t.Fatalf("Unsetenv: %v", err)
		}

		_, err := Parse([]string{"validate"})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("initはPATHをshelf rootとして使う", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")

		command, err := Parse([]string{"init", "new-shelf"})
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if command.Kind != application.Init || command.ShelfRoot != "new-shelf" {
			t.Fatalf("command = %#v", command)
		}
	})

	t.Run("newはタイトルと読了日を読む", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")

		command, err := Parse([]string{"new", "--read-date", "2024-01-02", "Some Book"})
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if command.Kind != application.New || command.Title != "Some Book" || command.ReadDate.Format("2006-01-02") != "2024-01-02" {
			t.Fatalf("command = %#v", command)
		}
	})

	t.Run("queryはwhere fragmentを読む", func(t *testing.T) {
		t.Setenv("SHELVIA_DIR", "env-shelf")

		command, err := Parse([]string{"query", "--where", "rating >= 90"})
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if command.Kind != application.Query || command.Where != "rating >= 90" {
			t.Fatalf("command = %#v", command)
		}
	})
}
