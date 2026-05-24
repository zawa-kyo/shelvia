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
	book, err := NewBook(validBookDraft(), testAllowedValues(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if book.Title().String() != "Some Book" {
		t.Fatalf("title = %q, want %q", book.Title().String(), "Some Book")
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
}

func TestNewBookAllowsEditionWithoutImprintWhenNotRequired(t *testing.T) {
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
}

func TestNewBookRejectsInvalidDraft(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*BookDraft)
		wantErr string
	}{
		{
			name: "missing title",
			mutate: func(draft *BookDraft) {
				draft.Title = ""
			},
			wantErr: "title",
		},
		{
			name: "unknown genre",
			mutate: func(draft *BookDraft) {
				draft.Genre = "Mystery"
			},
			wantErr: "unknown genre",
		},
		{
			name: "unknown publisher",
			mutate: func(draft *BookDraft) {
				draft.Publisher = "Unknown Publisher"
			},
			wantErr: "unknown publisher",
		},
		{
			name: "imprint without edition",
			mutate: func(draft *BookDraft) {
				draft.Edition = ""
			},
			wantErr: "imprint requires edition",
		},
		{
			name: "edition requires imprint",
			mutate: func(draft *BookDraft) {
				draft.Imprint = ""
			},
			wantErr: "imprint is required",
		},
		{
			name: "unknown edition",
			mutate: func(draft *BookDraft) {
				draft.Edition = "Unknown Edition"
			},
			wantErr: "unknown edition",
		},
		{
			name: "unknown imprint",
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
