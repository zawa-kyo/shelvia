package domain

import (
	"testing"
	"time"
)

func TestNewRating(t *testing.T) {
	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{name: "最小値の0を受け入れる", value: 0},
		{name: "中間値を受け入れる", value: 80},
		{name: "最大値の100を受け入れる", value: 100},
		{name: "0未満は拒否する", value: -1, wantErr: true},
		{name: "100超過は拒否する", value: 101, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRating(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Int() != tt.value {
				t.Fatalf("rating = %d, want %d", got.Int(), tt.value)
			}
		})
	}
}

func TestNewTitle(t *testing.T) {
	t.Run("前後の空白を取り除く", func(t *testing.T) {
		got, err := NewTitle("  Some Book  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.String() != "Some Book" {
			t.Fatalf("title = %q, want %q", got.String(), "Some Book")
		}
	})

	t.Run("空文字は拒否する", func(t *testing.T) {
		if _, err := NewTitle(" "); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestNewAuthor(t *testing.T) {
	t.Run("前後の空白を取り除く", func(t *testing.T) {
		got, err := NewAuthor("  Some Author  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.String() != "Some Author" {
			t.Fatalf("author = %q, want %q", got.String(), "Some Author")
		}
	})

	t.Run("空文字は拒否する", func(t *testing.T) {
		if _, err := NewAuthor(" "); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestNewReadDate(t *testing.T) {
	t.Run("時刻を落としてUTCの日付に正規化する", func(t *testing.T) {
		got, err := NewReadDate(time.Date(2024, 1, 2, 15, 4, 5, 0, time.Local))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.String() != "2024-01-02" {
			t.Fatalf("read date = %q, want %q", got.String(), "2024-01-02")
		}
		want := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
		if !got.Time().Equal(want) {
			t.Fatalf("read date time = %v, want %v", got.Time(), want)
		}
	})

	t.Run("ゼロ値は拒否する", func(t *testing.T) {
		if _, err := NewReadDate(time.Time{}); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestNewGenre(t *testing.T) {
	t.Run("前後の空白を取り除く", func(t *testing.T) {
		got, err := NewGenre("  Novel  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.String() != "Novel" {
			t.Fatalf("genre = %q, want %q", got.String(), "Novel")
		}
	})

	t.Run("空文字は拒否する", func(t *testing.T) {
		if _, err := NewGenre(" "); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestNewPublisher(t *testing.T) {
	t.Run("前後の空白を取り除く", func(t *testing.T) {
		got, err := NewPublisher("  Example Publisher  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.String() != "Example Publisher" {
			t.Fatalf("publisher = %q, want %q", got.String(), "Example Publisher")
		}
	})

	t.Run("空文字は拒否する", func(t *testing.T) {
		if _, err := NewPublisher(" "); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestNewEdition(t *testing.T) {
	t.Run("前後の空白を取り除く", func(t *testing.T) {
		got, err := NewEdition("  Paperback  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.IsSpecified() {
			t.Fatal("edition should be specified")
		}
		if got.String() != "Paperback" {
			t.Fatalf("edition = %q, want %q", got.String(), "Paperback")
		}
	})

	t.Run("空文字は未指定として扱う", func(t *testing.T) {
		got, err := NewEdition(" ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.IsSpecified() {
			t.Fatal("edition should be unspecified")
		}
		if got.String() != "" {
			t.Fatalf("edition = %q, want empty", got.String())
		}
	})
}

func TestNewImprint(t *testing.T) {
	t.Run("前後の空白を取り除く", func(t *testing.T) {
		got, err := NewImprint("  Example Paperback  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.IsSpecified() {
			t.Fatal("imprint should be specified")
		}
		if got.String() != "Example Paperback" {
			t.Fatalf("imprint = %q, want %q", got.String(), "Example Paperback")
		}
	})

	t.Run("空文字は未指定として扱う", func(t *testing.T) {
		got, err := NewImprint(" ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.IsSpecified() {
			t.Fatal("imprint should be unspecified")
		}
		if got.String() != "" {
			t.Fatalf("imprint = %q, want empty", got.String())
		}
	})
}

func TestNewOptionalText(t *testing.T) {
	t.Run("空文字は未指定として扱う", func(t *testing.T) {
		empty := NewOptionalText(" ")
		if empty.IsSpecified() {
			t.Fatal("blank optional text should be unspecified")
		}
		if empty.String() != "" {
			t.Fatalf("optional text = %q, want empty", empty.String())
		}
	})

	t.Run("前後の空白を取り除く", func(t *testing.T) {
		text := NewOptionalText("  note  ")
		if !text.IsSpecified() {
			t.Fatal("non-blank optional text should be specified")
		}
		if text.String() != "note" {
			t.Fatalf("optional text = %q, want %q", text.String(), "note")
		}
	})
}

func TestNewFilePath(t *testing.T) {
	t.Run("前後の空白を取り除く", func(t *testing.T) {
		got := NewFilePath("  books/example.toml  ")
		if got.String() != "books/example.toml" {
			t.Fatalf("file path = %q, want %q", got.String(), "books/example.toml")
		}
	})

	t.Run("空文字も診断用の未設定パスとして保持できる", func(t *testing.T) {
		got := NewFilePath(" ")
		if got.String() != "" {
			t.Fatalf("file path = %q, want empty", got.String())
		}
	})
}
