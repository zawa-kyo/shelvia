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
		t.Fatalf("読書記録に使う候補を準備できませんでした: %v", err)
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

func TestBookRecord(t *testing.T) {
	t.Run("必要な項目がそろっていれば本を記録できる", func(t *testing.T) {
		book, err := NewBook(validBookDraft(), testAllowedValues(t))
		if err != nil {
			t.Fatalf("本を記録できませんでした: %v", err)
		}

		if book.Title().String() != "Some Book" {
			t.Fatalf("title = %q, want %q", book.Title().String(), "Some Book")
		}
		if book.Author().String() != "Some Author" {
			t.Fatalf("author = %q, want %q", book.Author().String(), "Some Author")
		}
		if book.Rating().Int() != 90 {
			t.Fatalf("評価 = %d, want %d", book.Rating().Int(), 90)
		}
		if book.ReadDate().String() != "2024-01-02" {
			t.Fatalf("読了日 = %q, want %q", book.ReadDate().String(), "2024-01-02")
		}
		if book.Genre().String() != "Novel" {
			t.Fatalf("ジャンル = %q, want %q", book.Genre().String(), "Novel")
		}
		if book.Publisher().String() != "Example Publisher" {
			t.Fatalf("出版社 = %q, want %q", book.Publisher().String(), "Example Publisher")
		}
		if !book.Edition().IsSpecified() {
			t.Fatal("判型が記録されていません")
		}
		if !book.Imprint().IsSpecified() {
			t.Fatal("レーベルが記録されていません")
		}
		if book.Series().IsSpecified() {
			t.Fatal("空のシリーズは未指定として扱われるべきです")
		}
		if book.Translator().IsSpecified() {
			t.Fatal("空の訳者は未指定として扱われるべきです")
		}
		if book.Summary().IsSpecified() {
			t.Fatal("空のひとことメモは未指定として扱われるべきです")
		}
		if book.Body().IsSpecified() {
			t.Fatal("空の感想本文は未指定として扱われるべきです")
		}
	})
}

func TestEditionAndImprint(t *testing.T) {
	t.Run("レーベルが不要な判型はレーベルなしで記録できる", func(t *testing.T) {
		draft := validBookDraft()
		draft.Edition = "Hardcover"
		draft.Imprint = ""

		book, err := NewBook(draft, testAllowedValues(t))
		if err != nil {
			t.Fatalf("本を記録できませんでした: %v", err)
		}
		if !book.Edition().IsSpecified() {
			t.Fatal("判型が記録されていません")
		}
		if book.Imprint().IsSpecified() {
			t.Fatal("レーベルは未指定として扱われるべきです")
		}
	})
}

func TestOptionalFields(t *testing.T) {
	t.Run("任意項目は空なら未指定として記録できる", func(t *testing.T) {
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
			t.Fatalf("本を記録できませんでした: %v", err)
		}
		if book.Edition().IsSpecified() {
			t.Fatal("判型は未指定として扱われるべきです")
		}
		if book.Imprint().IsSpecified() {
			t.Fatal("レーベルは未指定として扱われるべきです")
		}
		if book.Series().IsSpecified() {
			t.Fatal("シリーズは未指定として扱われるべきです")
		}
		if book.Translator().IsSpecified() {
			t.Fatal("訳者は未指定として扱われるべきです")
		}
		if book.Summary().IsSpecified() {
			t.Fatal("ひとことメモは未指定として扱われるべきです")
		}
		if book.Body().IsSpecified() {
			t.Fatal("感想本文は未指定として扱われるべきです")
		}
		if book.FilePath().String() != "" {
			t.Fatalf("file path = %q, want empty", book.FilePath().String())
		}
	})

	t.Run("任意項目に書いた内容は読書記録に残る", func(t *testing.T) {
		draft := validBookDraft()
		draft.Series = "Series Name"
		draft.Translator = "Some Translator"
		draft.Summary = "短い感想"
		draft.Body = "長い感想"
		draft.FilePath = "books/some-book.toml"

		book, err := NewBook(draft, testAllowedValues(t))
		if err != nil {
			t.Fatalf("本を記録できませんでした: %v", err)
		}
		if book.Series().String() != "Series Name" {
			t.Fatalf("シリーズ = %q, want %q", book.Series().String(), "Series Name")
		}
		if book.Translator().String() != "Some Translator" {
			t.Fatalf("訳者 = %q, want %q", book.Translator().String(), "Some Translator")
		}
		if book.Summary().String() != "短い感想" {
			t.Fatalf("ひとことメモ = %q, want %q", book.Summary().String(), "短い感想")
		}
		if book.Body().String() != "長い感想" {
			t.Fatalf("感想本文 = %q, want %q", book.Body().String(), "長い感想")
		}
		if book.FilePath().String() != "books/some-book.toml" {
			t.Fatalf("記録元ファイル = %q, want %q", book.FilePath().String(), "books/some-book.toml")
		}
	})
}

func TestInvalidBookRecord(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*BookDraft)
		wantErr string
	}{
		{
			name: "題名のない本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Title = ""
			},
			wantErr: "title",
		},
		{
			name: "著者名のない本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Author = ""
			},
			wantErr: "author",
		},
		{
			name: "評価が100点を超える本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Rating = 101
			},
			wantErr: "rating",
		},
		{
			name: "読了日のない本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.ReadDate = time.Time{}
			},
			wantErr: "read_date",
		},
		{
			name: "知らないジャンルの本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Genre = "Mystery"
			},
			wantErr: "unknown genre",
		},
		{
			name: "ジャンルのない本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Genre = ""
			},
			wantErr: "genre",
		},
		{
			name: "知らない出版社の本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Publisher = "Unknown Publisher"
			},
			wantErr: "unknown publisher",
		},
		{
			name: "出版社のない本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Publisher = ""
			},
			wantErr: "publisher",
		},
		{
			name: "判型なしでレーベルだけある本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Edition = ""
			},
			wantErr: "imprint requires edition",
		},
		{
			name: "レーベルが必須の判型ではレーベルなしで記録できない",
			mutate: func(draft *BookDraft) {
				draft.Imprint = ""
			},
			wantErr: "imprint is required",
		},
		{
			name: "知らない判型の本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Edition = "Unknown Edition"
			},
			wantErr: "unknown edition",
		},
		{
			name: "知らないレーベルの本は記録できない",
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
				t.Fatal("記録できない本を受け入れてしまいました")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}
