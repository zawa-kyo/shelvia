package presentation

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zawa-kyo/shelvia/internal/application"
)

func TestWrite(t *testing.T) {
	t.Run("messageを表示する", func(t *testing.T) {
		var out bytes.Buffer
		output := application.Output{Message: "Validated 1 book, 1 config file."}

		Write(&out, output)

		require.Equal(t, "Validated 1 book, 1 config file.\n", out.String())
	})

	t.Run("tableを表示する", func(t *testing.T) {
		var out bytes.Buffer
		output := application.Output{Table: application.Table{
			Headers: []string{"read_date", "rating", "title"},
			Rows:    [][]string{{"2024-01-02", "90", "Some Book"}},
		}}

		Write(&out, output)

		got := out.String()
		for _, want := range []string{"Shelvia books", "1 row(s)", "read_date", "rating", "title", "Some Book"} {
			require.Contains(t, got, want)
		}
	})
}

func TestWriteError(t *testing.T) {
	t.Run("error messageを表示する", func(t *testing.T) {
		var out bytes.Buffer
		err := errString("Some Book.toml:5: unknown genre")

		WriteError(&out, err)

		require.Equal(t, "Some Book.toml:5: unknown genre\n", out.String())
	})
}

type errString string

func (err errString) Error() string {
	return string(err)
}
