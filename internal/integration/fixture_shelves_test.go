package integration_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/zawa-kyo/shelvia/internal/adapter/localfs"
	"github.com/zawa-kyo/shelvia/internal/adapter/queryengine"
	"github.com/zawa-kyo/shelvia/internal/application"
)

func TestFixtureShelves(t *testing.T) {
	t.Run("valid minimal shelfを検証できる", func(t *testing.T) {
		service := application.NewService(localfs.Repository{}, queryengine.Store{})

		output, err := service.Run(application.Command{
			Kind:      application.Validate,
			ShelfRoot: fixturePath(t, "valid", "minimal"),
		})
		if err != nil {
			t.Fatalf("Validate: %v", err)
		}
		if output.Message != "Validated 2 books, 1 config file." {
			t.Fatalf("message = %q", output.Message)
		}
	})

	tests := []struct {
		name string
		want string
	}{
		{name: "nested-config", want: "misplaced config.toml"},
		{name: "missing-required", want: "title"},
		{name: "unknown-config-value", want: "unknown genre"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"は検証エラーになる", func(t *testing.T) {
			service := application.NewService(localfs.Repository{}, queryengine.Store{})

			_, err := service.Run(application.Command{
				Kind:      application.Validate,
				ShelfRoot: fixturePath(t, "invalid", tt.name),
			})
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.want)
			}
		})
	}
}

func fixturePath(t *testing.T, parts ...string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Join(filepath.Dir(file), "..", "..", "fixtures", "shelves")
	all := append([]string{root}, parts...)
	return filepath.Clean(filepath.Join(all...))
}
