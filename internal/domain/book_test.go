package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err, "読書記録に使う候補を準備できませんでした")
	return values
}

func validBookDraft() BookDraft {
	return BookDraft{
		Title:     "Some Book",
		Author:    "Some Author",
		Rating:    intPointer(90),
		ReadDate:  time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		Genre:     "Novel",
		Publisher: "Example Publisher",
		Edition:   "Paperback",
		Imprint:   "Example Paperback",
		Series:    "  ",
	}
}

func intPointer(value int) *int {
	return &value
}

func TestBookRecord(t *testing.T) {
	t.Run("必要な項目がそろっていれば本を記録できる", func(t *testing.T) {
		draft := validBookDraft()
		allowed := testAllowedValues(t)

		book, err := NewBook(draft, allowed)

		require.NoError(t, err, "本を記録できませんでした")

		require.Equal(t, "Some Book", book.Title().String())
		require.Equal(t, "Some Author", book.Author().String())
		require.Equal(t, 90, book.Rating().Int())
		require.Equal(t, "2024-01-02", book.ReadDate().String())
		require.Equal(t, "Novel", book.Genre().String())
		require.Equal(t, "Example Publisher", book.Publisher().String())
		require.True(t, book.Edition().IsSpecified(), "判型が記録されていません")
		require.True(t, book.Imprint().IsSpecified(), "レーベルが記録されていません")
		require.False(t, book.Series().IsSpecified(), "空のシリーズは未指定として扱われるべきです")
		require.False(t, book.Translator().IsSpecified(), "空の訳者は未指定として扱われるべきです")
		require.False(t, book.Summary().IsSpecified(), "空のひとことメモは未指定として扱われるべきです")
		require.False(t, book.Body().IsSpecified(), "空の感想本文は未指定として扱われるべきです")
	})

	t.Run("評価0点の本を記録できる", func(t *testing.T) {
		draft := validBookDraft()
		draft.Rating = intPointer(0)
		allowed := testAllowedValues(t)

		book, err := NewBook(draft, allowed)

		require.NoError(t, err, "本を記録できませんでした")
		require.Equal(t, 0, book.Rating().Int())
	})
}

func TestEditionAndImprint(t *testing.T) {
	t.Run("レーベルが不要な判型はレーベルなしで記録できる", func(t *testing.T) {
		draft := validBookDraft()
		draft.Edition = "Hardcover"
		draft.Imprint = ""
		allowed := testAllowedValues(t)

		book, err := NewBook(draft, allowed)

		require.NoError(t, err, "本を記録できませんでした")
		require.True(t, book.Edition().IsSpecified(), "判型が記録されていません")
		require.False(t, book.Imprint().IsSpecified(), "レーベルは未指定として扱われるべきです")
	})

	t.Run("レーベルが不要な判型にも許可されたレーベルを記録できる", func(t *testing.T) {
		draft := validBookDraft()
		draft.Edition = "Hardcover"
		allowed := testAllowedValues(t)

		book, err := NewBook(draft, allowed)

		require.NoError(t, err, "本を記録できませんでした")
		require.Equal(t, "Hardcover", book.Edition().String())
		require.Equal(t, "Example Paperback", book.Imprint().String())
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
		allowed := testAllowedValues(t)

		book, err := NewBook(draft, allowed)

		require.NoError(t, err, "本を記録できませんでした")
		require.False(t, book.Edition().IsSpecified(), "判型は未指定として扱われるべきです")
		require.False(t, book.Imprint().IsSpecified(), "レーベルは未指定として扱われるべきです")
		require.False(t, book.Series().IsSpecified(), "シリーズは未指定として扱われるべきです")
		require.False(t, book.Translator().IsSpecified(), "訳者は未指定として扱われるべきです")
		require.False(t, book.Summary().IsSpecified(), "ひとことメモは未指定として扱われるべきです")
		require.False(t, book.Body().IsSpecified(), "感想本文は未指定として扱われるべきです")
		require.Empty(t, book.FilePath().String())
	})

	t.Run("任意項目に書いた内容は読書記録に残る", func(t *testing.T) {
		draft := validBookDraft()
		draft.Series = "Series Name"
		draft.Translator = "Some Translator"
		draft.Summary = "短い感想"
		draft.Body = "長い感想"
		draft.FilePath = "books/some-book.toml"
		allowed := testAllowedValues(t)

		book, err := NewBook(draft, allowed)

		require.NoError(t, err, "本を記録できませんでした")
		require.Equal(t, "Series Name", book.Series().String())
		require.Equal(t, "Some Translator", book.Translator().String())
		require.Equal(t, "短い感想", book.Summary().String())
		require.Equal(t, "長い感想", book.Body().String())
		require.Equal(t, "books/some-book.toml", book.FilePath().String())
	})
}

func TestInvalidBookRecord(t *testing.T) {
	t.Run("評価のない本は記録できない", func(t *testing.T) {
		draft := validBookDraft()
		draft.Rating = nil
		allowed := testAllowedValues(t)

		_, err := NewBook(draft, allowed)

		require.Equal(t, ValidationError{Field: "rating", Reason: "required"}, err)
	})

	tests := []struct {
		name       string
		mutate     func(*BookDraft)
		wantField  string
		wantReason string
	}{
		{
			name: "題名のない本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Title = ""
			},
			wantField: "title",
		},
		{
			name: "著者名のない本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Author = ""
			},
			wantField: "author",
		},
		{
			name: "評価が100点を超える本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Rating = intPointer(101)
			},
			wantField: "rating",
		},
		{
			name: "読了日のない本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.ReadDate = time.Time{}
			},
			wantField: "read_date",
		},
		{
			name: "知らないジャンルの本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Genre = "Mystery"
			},
			wantField:  "genre",
			wantReason: "unknown genre",
		},
		{
			name: "ジャンルのない本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Genre = ""
			},
			wantField: "genre",
		},
		{
			name: "知らない出版社の本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Publisher = "Unknown Publisher"
			},
			wantField:  "publisher",
			wantReason: "unknown publisher",
		},
		{
			name: "出版社のない本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Publisher = ""
			},
			wantField: "publisher",
		},
		{
			name: "判型なしでレーベルだけある本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Edition = ""
			},
			wantField:  "imprint",
			wantReason: "imprint requires edition",
		},
		{
			name: "レーベルが必須の判型ではレーベルなしで記録できない",
			mutate: func(draft *BookDraft) {
				draft.Imprint = ""
			},
			wantField:  "imprint",
			wantReason: "imprint is required",
		},
		{
			name: "知らない判型の本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Edition = "Unknown Edition"
			},
			wantField:  "edition",
			wantReason: "unknown edition",
		},
		{
			name: "知らないレーベルの本は記録できない",
			mutate: func(draft *BookDraft) {
				draft.Imprint = "Unknown Imprint"
			},
			wantField:  "imprint",
			wantReason: "unknown imprint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			draft := validBookDraft()
			tt.mutate(&draft)
			allowed := testAllowedValues(t)

			_, err := NewBook(draft, allowed)

			require.Error(t, err, "記録できない本を受け入れてしまいました")
			var validationErr ValidationError
			require.ErrorAs(t, err, &validationErr)
			require.Equal(t, tt.wantField, validationErr.Field)
			if tt.wantReason != "" {
				require.Contains(t, validationErr.Reason, tt.wantReason)
			}
		})
	}
}
