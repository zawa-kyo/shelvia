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

		err := Write(&out, output)

		require.NoError(t, err)
		require.Equal(t, "Validated 1 book, 1 config file.\n", out.String())
	})

	t.Run("tableを表示する", func(t *testing.T) {
		var out bytes.Buffer
		result := tableData{
			Headers: []string{"read_date", "rating", "title"},
			Rows: [][]string{
				{"2024-01-02", "90", "Some Book"},
				{"2024-02-03", "", "1234567890123456789012345"},
			},
		}

		err := writeTable(&out, result, false)

		require.NoError(t, err)
		got := out.String()
		for _, want := range []string{
			"Shelvia books",
			"2 row(s)",
			"read_date",
			"rating",
			"title",
			"2024-01-02",
			"90",
			"Some Book",
			"2024-02-03",
			"-",
			"123456789012345678901...",
		} {
			require.Contains(t, got, want)
		}
		require.NotContains(t, got, "1234567890123456789012345")
	})

	t.Run("headerがない出力には何も表示しない", func(t *testing.T) {
		var out bytes.Buffer

		err := Write(&out, application.Output{})

		require.NoError(t, err)
		require.Empty(t, out.String())
	})

	t.Run("messageがある場合はtableより優先して表示する", func(t *testing.T) {
		var out bytes.Buffer
		output := application.Output{Message: "Done.", Kind: application.List}

		err := Write(&out, output)

		require.NoError(t, err)
		require.Equal(t, "Done.\n", out.String())
	})
}

func TestWriteError(t *testing.T) {
	t.Run("error messageを表示する", func(t *testing.T) {
		var out bytes.Buffer
		err := errString("Some Book.toml:5: unknown genre")

		WriteError(&out, err)

		require.Equal(t, "Some Book.toml:5: unknown genre\n", out.String())
	})

	t.Run("errorがなければ何も表示しない", func(t *testing.T) {
		var out bytes.Buffer

		WriteError(&out, nil)

		require.Empty(t, out.String())
	})
}

type errString string

func (err errString) Error() string {
	return string(err)
}
