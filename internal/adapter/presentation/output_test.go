package presentation

import (
	"bytes"
	"strings"
	"testing"

	"github.com/zawa-kyo/shelvia/internal/application"
)

func TestWrite(t *testing.T) {
	t.Run("messageを表示する", func(t *testing.T) {
		var out bytes.Buffer

		Write(&out, application.Output{Message: "Validated 1 book, 1 config file."})

		if out.String() != "Validated 1 book, 1 config file.\n" {
			t.Fatalf("output = %q", out.String())
		}
	})

	t.Run("tableを表示する", func(t *testing.T) {
		var out bytes.Buffer
		output := application.Output{Table: application.Table{
			Headers: []string{"read_date", "rating", "title"},
			Rows:    [][]string{{"2024-01-02", "90", "Some Book"}},
		}}

		Write(&out, output)

		got := out.String()
		for _, want := range []string{"read_date", "rating", "title", "Some Book"} {
			if !strings.Contains(got, want) {
				t.Fatalf("output = %q, want to contain %q", got, want)
			}
		}
	})
}

func TestWriteError(t *testing.T) {
	t.Run("error messageを表示する", func(t *testing.T) {
		var out bytes.Buffer

		WriteError(&out, errString("Some Book.toml:5: unknown genre"))

		if out.String() != "Some Book.toml:5: unknown genre\n" {
			t.Fatalf("output = %q", out.String())
		}
	})
}

type errString string

func (err errString) Error() string {
	return string(err)
}
