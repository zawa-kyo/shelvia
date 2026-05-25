package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRating(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{name: "0点も記録できる", value: 0},
		{name: "読後感を点数で記録できる", value: 80},
		{name: "100点も記録できる", value: 100},
		{name: "0点未満は記録できない", value: -1, wantErr: true},
		{name: "100点を超える評価は記録できない", value: 101, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRating(tt.value)
			if tt.wantErr {
				require.Error(t, err, "記録できない評価を受け入れてしまいました")
				return
			}
			require.NoError(t, err, "評価を記録できませんでした")
			require.Equal(t, tt.value, got.Int())
		})
	}
}

func TestTitle(t *testing.T) {
	t.Run("前後の空白は題名に含めない", func(t *testing.T) {
		got, err := NewTitle("  Some Book  ")
		require.NoError(t, err, "題名を記録できませんでした")
		require.Equal(t, "Some Book", got.String())
	})

	t.Run("題名のない本は記録できない", func(t *testing.T) {
		_, err := NewTitle(" ")
		require.Error(t, err, "題名のない本を受け入れてしまいました")
	})
}

func TestAuthor(t *testing.T) {
	t.Run("前後の空白は著者名に含めない", func(t *testing.T) {
		got, err := NewAuthor("  Some Author  ")
		require.NoError(t, err, "著者名を記録できませんでした")
		require.Equal(t, "Some Author", got.String())
	})

	t.Run("著者名のない本は記録できない", func(t *testing.T) {
		_, err := NewAuthor(" ")
		require.Error(t, err, "著者名のない本を受け入れてしまいました")
	})
}

func TestReadDate(t *testing.T) {
	t.Run("時刻ではなく読んだ日だけを記録する", func(t *testing.T) {
		got, err := NewReadDate(time.Date(2024, 1, 2, 15, 4, 5, 0, time.Local))
		require.NoError(t, err, "読了日を記録できませんでした")
		require.Equal(t, "2024-01-02", got.String())

		want := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		require.True(t, got.Time().Equal(want), "読了日は日付だけで保存されます")
	})

	t.Run("読了日のない本は記録できない", func(t *testing.T) {
		_, err := NewReadDate(time.Time{})
		require.Error(t, err, "読了日のない本を受け入れてしまいました")
	})
}

func TestGenre(t *testing.T) {
	t.Run("前後の空白はジャンル名に含めない", func(t *testing.T) {
		got, err := NewGenre("  Novel  ")
		require.NoError(t, err, "ジャンルを記録できませんでした")
		require.Equal(t, "Novel", got.String())
	})

	t.Run("ジャンルのない本は記録できない", func(t *testing.T) {
		_, err := NewGenre(" ")
		require.Error(t, err, "ジャンルのない本を受け入れてしまいました")
	})
}

func TestPublisher(t *testing.T) {
	t.Run("前後の空白は出版社名に含めない", func(t *testing.T) {
		got, err := NewPublisher("  Example Publisher  ")
		require.NoError(t, err, "出版社を記録できませんでした")
		require.Equal(t, "Example Publisher", got.String())
	})

	t.Run("出版社のない本は記録できない", func(t *testing.T) {
		_, err := NewPublisher(" ")
		require.Error(t, err, "出版社のない本を受け入れてしまいました")
	})
}

func TestEdition(t *testing.T) {
	t.Run("前後の空白は判型に含めない", func(t *testing.T) {
		got, err := NewEdition("  Paperback  ")
		require.NoError(t, err, "判型を記録できませんでした")
		require.True(t, got.IsSpecified(), "判型が記録されていません")
		require.Equal(t, "Paperback", got.String())
	})

	t.Run("判型は空なら未指定として扱う", func(t *testing.T) {
		got, err := NewEdition(" ")
		require.NoError(t, err, "判型を扱えませんでした")
		require.False(t, got.IsSpecified(), "空の判型は未指定として扱われるべきです")
		require.Empty(t, got.String())
	})
}

func TestImprint(t *testing.T) {
	t.Run("前後の空白はレーベル名に含めない", func(t *testing.T) {
		got, err := NewImprint("  Example Paperback  ")
		require.NoError(t, err, "レーベルを記録できませんでした")
		require.True(t, got.IsSpecified(), "レーベルが記録されていません")
		require.Equal(t, "Example Paperback", got.String())
	})

	t.Run("レーベルは空なら未指定として扱う", func(t *testing.T) {
		got, err := NewImprint(" ")
		require.NoError(t, err, "レーベルを扱えませんでした")
		require.False(t, got.IsSpecified(), "空のレーベルは未指定として扱われるべきです")
		require.Empty(t, got.String())
	})
}

func TestOptionalText(t *testing.T) {
	t.Run("空のメモは未指定として扱う", func(t *testing.T) {
		empty := NewOptionalText(" ")
		require.False(t, empty.IsSpecified(), "空のメモは未指定として扱われるべきです")
		require.Empty(t, empty.String())
	})

	t.Run("前後の空白はメモ本文に含めない", func(t *testing.T) {
		text := NewOptionalText("  note  ")
		require.True(t, text.IsSpecified(), "書かれたメモが記録されていません")
		require.Equal(t, "note", text.String())
	})
}

func TestFilePath(t *testing.T) {
	t.Run("前後の空白はファイル名に含めない", func(t *testing.T) {
		got := NewFilePath("  books/example.toml  ")
		require.Equal(t, "books/example.toml", got.String())
	})

	t.Run("ファイル名が分からない記録も扱える", func(t *testing.T) {
		got := NewFilePath(" ")
		require.Empty(t, got.String())
	})
}
