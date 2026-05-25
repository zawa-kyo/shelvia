package domain

import (
	"testing"
	"time"
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
				if err == nil {
					t.Fatal("記録できない評価を受け入れてしまいました")
				}
				return
			}
			if err != nil {
				t.Fatalf("評価を記録できませんでした: %v", err)
			}
			if got.Int() != tt.value {
				t.Fatalf("評価 = %d, want %d", got.Int(), tt.value)
			}
		})
	}
}

func TestTitle(t *testing.T) {
	t.Run("前後の空白は題名に含めない", func(t *testing.T) {
		got, err := NewTitle("  Some Book  ")
		if err != nil {
			t.Fatalf("題名を記録できませんでした: %v", err)
		}
		if got.String() != "Some Book" {
			t.Fatalf("title = %q, want %q", got.String(), "Some Book")
		}
	})

	t.Run("題名のない本は記録できない", func(t *testing.T) {
		if _, err := NewTitle(" "); err == nil {
			t.Fatal("題名のない本を受け入れてしまいました")
		}
	})
}

func TestAuthor(t *testing.T) {
	t.Run("前後の空白は著者名に含めない", func(t *testing.T) {
		got, err := NewAuthor("  Some Author  ")
		if err != nil {
			t.Fatalf("著者名を記録できませんでした: %v", err)
		}
		if got.String() != "Some Author" {
			t.Fatalf("author = %q, want %q", got.String(), "Some Author")
		}
	})

	t.Run("著者名のない本は記録できない", func(t *testing.T) {
		if _, err := NewAuthor(" "); err == nil {
			t.Fatal("著者名のない本を受け入れてしまいました")
		}
	})
}

func TestReadDate(t *testing.T) {
	t.Run("時刻ではなく読んだ日だけを記録する", func(t *testing.T) {
		got, err := NewReadDate(time.Date(2024, 1, 2, 15, 4, 5, 0, time.Local))
		if err != nil {
			t.Fatalf("読了日を記録できませんでした: %v", err)
		}
		if got.String() != "2024-01-02" {
			t.Fatalf("読了日 = %q, want %q", got.String(), "2024-01-02")
		}
		want := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		if !got.Time().Equal(want) {
			t.Fatalf("読了日 = %v, want %v", got.Time(), want)
		}
	})

	t.Run("読了日のない本は記録できない", func(t *testing.T) {
		if _, err := NewReadDate(time.Time{}); err == nil {
			t.Fatal("読了日のない本を受け入れてしまいました")
		}
	})
}

func TestGenre(t *testing.T) {
	t.Run("前後の空白はジャンル名に含めない", func(t *testing.T) {
		got, err := NewGenre("  Novel  ")
		if err != nil {
			t.Fatalf("ジャンルを記録できませんでした: %v", err)
		}
		if got.String() != "Novel" {
			t.Fatalf("ジャンル = %q, want %q", got.String(), "Novel")
		}
	})

	t.Run("ジャンルのない本は記録できない", func(t *testing.T) {
		if _, err := NewGenre(" "); err == nil {
			t.Fatal("ジャンルのない本を受け入れてしまいました")
		}
	})
}

func TestPublisher(t *testing.T) {
	t.Run("前後の空白は出版社名に含めない", func(t *testing.T) {
		got, err := NewPublisher("  Example Publisher  ")
		if err != nil {
			t.Fatalf("出版社を記録できませんでした: %v", err)
		}
		if got.String() != "Example Publisher" {
			t.Fatalf("出版社 = %q, want %q", got.String(), "Example Publisher")
		}
	})

	t.Run("出版社のない本は記録できない", func(t *testing.T) {
		if _, err := NewPublisher(" "); err == nil {
			t.Fatal("出版社のない本を受け入れてしまいました")
		}
	})
}

func TestEdition(t *testing.T) {
	t.Run("前後の空白は判型に含めない", func(t *testing.T) {
		got, err := NewEdition("  Paperback  ")
		if err != nil {
			t.Fatalf("判型を記録できませんでした: %v", err)
		}
		if !got.IsSpecified() {
			t.Fatal("判型が記録されていません")
		}
		if got.String() != "Paperback" {
			t.Fatalf("判型 = %q, want %q", got.String(), "Paperback")
		}
	})

	t.Run("判型は空なら未指定として扱う", func(t *testing.T) {
		got, err := NewEdition(" ")
		if err != nil {
			t.Fatalf("判型を扱えませんでした: %v", err)
		}
		if got.IsSpecified() {
			t.Fatal("空の判型は未指定として扱われるべきです")
		}
		if got.String() != "" {
			t.Fatalf("判型 = %q, want empty", got.String())
		}
	})
}

func TestImprint(t *testing.T) {
	t.Run("前後の空白はレーベル名に含めない", func(t *testing.T) {
		got, err := NewImprint("  Example Paperback  ")
		if err != nil {
			t.Fatalf("レーベルを記録できませんでした: %v", err)
		}
		if !got.IsSpecified() {
			t.Fatal("レーベルが記録されていません")
		}
		if got.String() != "Example Paperback" {
			t.Fatalf("レーベル = %q, want %q", got.String(), "Example Paperback")
		}
	})

	t.Run("レーベルは空なら未指定として扱う", func(t *testing.T) {
		got, err := NewImprint(" ")
		if err != nil {
			t.Fatalf("レーベルを扱えませんでした: %v", err)
		}
		if got.IsSpecified() {
			t.Fatal("空のレーベルは未指定として扱われるべきです")
		}
		if got.String() != "" {
			t.Fatalf("レーベル = %q, want empty", got.String())
		}
	})
}

func TestOptionalText(t *testing.T) {
	t.Run("空のメモは未指定として扱う", func(t *testing.T) {
		empty := NewOptionalText(" ")
		if empty.IsSpecified() {
			t.Fatal("空のメモは未指定として扱われるべきです")
		}
		if empty.String() != "" {
			t.Fatalf("メモ = %q, want empty", empty.String())
		}
	})

	t.Run("前後の空白はメモ本文に含めない", func(t *testing.T) {
		text := NewOptionalText("  note  ")
		if !text.IsSpecified() {
			t.Fatal("書かれたメモが記録されていません")
		}
		if text.String() != "note" {
			t.Fatalf("メモ = %q, want %q", text.String(), "note")
		}
	})
}

func TestFilePath(t *testing.T) {
	t.Run("前後の空白はファイル名に含めない", func(t *testing.T) {
		got := NewFilePath("  books/example.toml  ")
		if got.String() != "books/example.toml" {
			t.Fatalf("file path = %q, want %q", got.String(), "books/example.toml")
		}
	})

	t.Run("ファイル名が分からない記録も扱える", func(t *testing.T) {
		got := NewFilePath(" ")
		if got.String() != "" {
			t.Fatalf("file path = %q, want empty", got.String())
		}
	})
}
