package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zawa-kyo/shelvia/internal/application"
)

func TestParseShelfRoot(t *testing.T) {
	t.Run("明示的なshelfが環境変数より優先される", func(t *testing.T) {
		t.Setenv(shelfEnv, "env-shelf")

		command, err := Parse([]string{"--shelf", "arg-shelf", "validate"})

		require.NoError(t, err)
		require.Equal(t, application.Command{Kind: application.Validate, ShelfRoot: "arg-shelf"}, command)
	})

	t.Run("shelf未指定なら環境変数を参照する", func(t *testing.T) {
		t.Setenv(shelfEnv, "env-shelf")

		command, err := Parse([]string{"list"})

		require.NoError(t, err)
		require.Equal(t, application.Command{Kind: application.List, ShelfRoot: "env-shelf"}, command)
	})

	t.Run("どちらもない場合はエラーにする", func(t *testing.T) {
		t.Setenv(shelfEnv, "")

		_, err := Parse([]string{"validate"})

		require.EqualError(t, err, "shelf root is not set; pass --shelf or set SHELVIA_DIR")
	})
}

func TestParseInit(t *testing.T) {
	t.Setenv(shelfEnv, "env-shelf")
	tests := []struct {
		name string
		args []string
		root string
	}{
		{name: "PATHをshelf rootとして使う", args: []string{"init", "new-shelf"}, root: "new-shelf"},
		{name: "PATH省略時はカレントディレクトリを使う", args: []string{"init"}, root: "."},
		{name: "PATH省略時は明示的なshelfを使う", args: []string{"--shelf", "arg-shelf", "init"}, root: "arg-shelf"},
		{name: "PATHは明示的なshelfより優先される", args: []string{"--shelf", "arg-shelf", "init", "new-shelf"}, root: "new-shelf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := Parse(tt.args)

			require.NoError(t, err)
			require.Equal(t, application.Command{Kind: application.Init, ShelfRoot: tt.root}, command)
		})
	}
}

func TestParseCommand(t *testing.T) {
	t.Setenv(shelfEnv, "shelf")
	readDate := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		args []string
		want application.Command
	}{
		{
			name: "addはタイトルと日付を読む",
			args: []string{"add", "Some Book", "--date", "2024-01-02"},
			want: application.Command{Kind: application.Add, ShelfRoot: "shelf", Title: "Some Book", ReadDate: readDate},
		},
		{
			name: "addは日付をタイトルより先に読める",
			args: []string{"add", "--date", "2024-01-02", "Some Book"},
			want: application.Command{Kind: application.Add, ShelfRoot: "shelf", Title: "Some Book", ReadDate: readDate},
		},
		{
			name: "validateを読む",
			args: []string{"validate"},
			want: application.Command{Kind: application.Validate, ShelfRoot: "shelf"},
		},
		{
			name: "listを読む",
			args: []string{"list"},
			want: application.Command{Kind: application.List, ShelfRoot: "shelf"},
		},
		{
			name: "searchはwhere fragmentを読む",
			args: []string{"search", "rating >= 90"},
			want: application.Command{Kind: application.Search, ShelfRoot: "shelf", Where: "rating >= 90"},
		},
		{
			name: "showはタイトルを読む",
			args: []string{"show", "Some Book"},
			want: application.Command{Kind: application.Show, ShelfRoot: "shelf", Title: "Some Book"},
		},
		{
			name: "pathはタイトルを読む",
			args: []string{"path", "Some Book"},
			want: application.Command{Kind: application.Path, ShelfRoot: "shelf", Title: "Some Book"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := Parse(tt.args)

			require.NoError(t, err)
			require.Equal(t, tt.want, command)
		})
	}
}

func TestParseRejectsInvalidArguments(t *testing.T) {
	t.Setenv(shelfEnv, "shelf")
	tests := []struct {
		name         string
		args         []string
		wantErr      string
		containsOnly bool
	}{
		{name: "commandがない", args: nil, wantErr: "missing command"},
		{name: "未知のglobal option", args: []string{"--unknown", "validate"}, wantErr: "unknown", containsOnly: true},
		{name: "initのPATHが複数ある", args: []string{"init", "one", "two"}, wantErr: "usage: shelvia init [PATH]"},
		{name: "addのタイトルがない", args: []string{"add", "--date", "2024-01-02"}, wantErr: "usage: shelvia add TITLE --date YYYY-MM-DD"},
		{name: "addの日付optionがない", args: []string{"add", "Some Book"}, wantErr: "add requires --date"},
		{name: "addの日付optionに値がない", args: []string{"add", "Some Book", "--date"}, wantErr: "add requires --date"},
		{name: "addの日付がISO形式でない", args: []string{"add", "Some Book", "--date", "2024/01/02"}, wantErr: "invalid --date: 2024/01/02"},
		{name: "addに未知のoptionがある", args: []string{"add", "Some Book", "--unknown", "value", "--date", "2024-01-02"}, wantErr: "unknown add option: --unknown"},
		{name: "addのタイトルが複数ある", args: []string{"add", "Some Book", "Other Book", "--date", "2024-01-02"}, wantErr: "usage: shelvia add TITLE --date YYYY-MM-DD"},
		{name: "validateに引数がある", args: []string{"validate", "extra"}, wantErr: "usage: shelvia validate"},
		{name: "listに引数がある", args: []string{"list", "extra"}, wantErr: "usage: shelvia list"},
		{name: "searchに条件がない", args: []string{"search"}, wantErr: "usage: shelvia search SQL_FRAGMENT"},
		{name: "showにタイトルがない", args: []string{"show"}, wantErr: "usage: shelvia show TITLE"},
		{name: "pathにタイトルがない", args: []string{"path"}, wantErr: "usage: shelvia path TITLE"},
		{name: "未知のcommand", args: []string{"query"}, wantErr: "unknown command: query"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.args)

			if tt.containsOnly {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.EqualError(t, err, tt.wantErr)
		})
	}
}
