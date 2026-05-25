package domain

import (
	"strings"
	"testing"
	"time"
)

func testAllowedValues(t *testing.T) AllowedValues {
	t.Helper()

	values, err := NewAllowedValues(AllowedValuesInput{
		Genres:     []string{"Novel", "Technical"},
		Publishers: []string{"Example Publisher"},
		Imprints:   []string{"Example Paperback"},
		Editions: []EditionInput{
			{Name: "Paperback", ImprintRequired: true},
			{Name: "Hardcover", ImprintRequired: false},
		},
	})
	if err != nil {
		t.Fatalf("NewAllowedValues: %v", err)
	}
	return values
}

func validBookDraft() BookDraft {
	return BookDraft{
		Title:     "Some Book",
		Author:    "Some Author",
		Rating:    90,
		ReadDate:  time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		Genre:     "Novel",
		Publisher: "Example Publisher",
		Edition:   "Paperback",
		Imprint:   "Example Paperback",
		Series:    "  ",
	}
}

func TestNewBook(t *testing.T) {
	t.Run("有効な入力から読書記録を作成する", func(t *testing.T) {
		book, err := NewBook(validBookDraft(), testAllowedValues(t))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if book.Title().String() != "Some Book" {
			t.Fatalf("title = %q, want %q", book.Title().String(), "Some Book")
		}
		if book.Author().String() != "Some Author" {
			t.Fatalf("author = %q, want %q", book.Author().String(), "Some Author")
		}
		if book.Rating().Int() != 90 {
			t.Fatalf("rating = %d, want %d", book.Rating().Int(), 90)
		}
		if book.ReadDate().String() != "2024-01-02" {
			t.Fatalf("read date = %q, want %q", book.ReadDate().String(), "2024-01-02")
		}
		if book.Genre().String() != "Novel" {
			t.Fatalf("genre = %q, want %q", book.Genre().String(), "Novel")
		}
		if book.Publisher().String() != "Example Publisher" {
			t.Fatalf("publisher = %q, want %q", book.Publisher().String(), "Example Publisher")
		}
		if !book.Edition().IsSpecified() {
			t.Fatal("edition should be specified")
		}
		if !book.Imprint().IsSpecified() {
			t.Fatal("imprint should be specified")
		}
		if book.Series().IsSpecified() {
			t.Fatal("blank series should be unspecified")
		}
		if book.Translator().IsSpecified() {
			t.Fatal("blank translator should be unspecified")
		}
		if book.Summary().IsSpecified() {
			t.Fatal("blank summary should be unspecified")
		}
		if book.Body().IsSpecified() {
			t.Fatal("blank body should be unspecified")
		}
	})
}

func TestNewBookAllowsEditionWithoutImprintWhenNotRequired(t *testing.T) {
	t.Run("印刷所が不要な版は印刷所なしで作成できる", func(t *testing.T) {
		draft := validBookDraft()
		draft.Edition = "Hardcover"
		draft.Imprint = ""

		book, err := NewBook(draft, testAllowedValues(t))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !book.Edition().IsSpecified() {
			t.Fatal("edition should be specified")
		}
		if book.Imprint().IsSpecified() {
			t.Fatal("imprint should be unspecified")
		}
	})
}

func TestNewBookAllowsOptionalFieldsToBeAbsent(t *testing.T) {
	t.Run("任意項目は空文字なら未指定として作成できる", func(t *testing.T) {
		draft := validBookDraft()
		draft.Edition = ""
		draft.Imprint = ""
		draft.Series = ""
		draft.Translator = ""
		draft.Summary = ""
		draft.Body = ""
		draft.FilePath = ""

		book, err := NewBook(draft, testAllowedValues(t))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if book.Edition().IsSpecified() {
			t.Fatal("edition should be unspecified")
		}
		if book.Imprint().IsSpecified() {
			t.Fatal("imprint should be unspecified")
		}
		if book.Series().IsSpecified() {
			t.Fatal("series should be unspecified")
		}
		if book.Translator().IsSpecified() {
			t.Fatal("translator should be unspecified")
		}
		if book.Summary().IsSpecified() {
			t.Fatal("summary should be unspecified")
		}
		if book.Body().IsSpecified() {
			t.Fatal("body should be unspecified")
		}
		if book.FilePath().String() != "" {
			t.Fatalf("file path = %q, want empty", book.FilePath().String())
		}
	})
}

func TestNewBookRejectsInvalidDraft(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*BookDraft)
		wantErr string
	}{
		{
			name: "タイトルがない入力は拒否する",
			mutate: func(draft *BookDraft) {
				draft.Title = ""
			},
			wantErr: "title",
		},
		{
			name: "著者がない入力は拒否する",
			mutate: func(draft *BookDraft) {
				draft.Author = ""
			},
			wantErr: "author",
		},
		{
			name: "評価が範囲外の入力は拒否する",
			mutate: func(draft *BookDraft) {
				draft.Rating = 101
			},
			wantErr: "rating",
		},
		{
			name: "読了日がない入力は拒否する",
			mutate: func(draft *BookDraft) {
				draft.ReadDate = time.Time{}
			},
			wantErr: "read_date",
		},
		{
			name: "ジャンルが設定値にない入力は拒否する",
			mutate: func(draft *BookDraft) {
				draft.Genre = "Mystery"
			},
			wantErr: "unknown genre",
		},
		{
			name: "ジャンルがない入力は拒否する",
			mutate: func(draft *BookDraft) {
				draft.Genre = ""
			},
			wantErr: "genre",
		},
		{
			name: "出版社が設定値にない入力は拒否する",
			mutate: func(draft *BookDraft) {
				draft.Publisher = "Unknown Publisher"
			},
			wantErr: "unknown publisher",
		},
		{
			name: "出版社がない入力は拒否する",
			mutate: func(draft *BookDraft) {
				draft.Publisher = ""
			},
			wantErr: "publisher",
		},
		{
			name: "版なしで印刷所だけある入力は拒否する",
			mutate: func(draft *BookDraft) {
				draft.Edition = ""
			},
			wantErr: "imprint requires edition",
		},
		{
			name: "印刷所が必須の版で印刷所がない入力は拒否する",
			mutate: func(draft *BookDraft) {
				draft.Imprint = ""
			},
			wantErr: "imprint is required",
		},
		{
			name: "版が設定値にない入力は拒否する",
			mutate: func(draft *BookDraft) {
				draft.Edition = "Unknown Edition"
			},
			wantErr: "unknown edition",
		},
		{
			name: "印刷所が設定値にない入力は拒否する",
			mutate: func(draft *BookDraft) {
				draft.Imprint = "Unknown Imprint"
			},
			wantErr: "unknown imprint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draft := validBookDraft()
			tt.mutate(&draft)

			_, err := NewBook(draft, testAllowedValues(t))
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}
