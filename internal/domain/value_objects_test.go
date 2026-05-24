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
		{name: "minimum", value: 0},
		{name: "middle", value: 80},
		{name: "maximum", value: 100},
		{name: "below minimum", value: -1, wantErr: true},
		{name: "above maximum", value: 101, wantErr: true},
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
	got, err := NewTitle("  Some Book  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.String() != "Some Book" {
		t.Fatalf("title = %q, want %q", got.String(), "Some Book")
	}

	if _, err := NewTitle(" "); err == nil {
		t.Fatal("expected error for blank title")
	}
}

func TestNewReadDate(t *testing.T) {
	got, err := NewReadDate(time.Date(2024, 1, 2, 15, 4, 5, 0, time.Local))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.String() != "2024-01-02" {
		t.Fatalf("read date = %q, want %q", got.String(), "2024-01-02")
	}

	if _, err := NewReadDate(time.Time{}); err == nil {
		t.Fatal("expected error for zero date")
	}
}

func TestNewOptionalText(t *testing.T) {
	empty := NewOptionalText(" ")
	if empty.IsSpecified() {
		t.Fatal("blank optional text should be unspecified")
	}

	text := NewOptionalText("  note  ")
	if !text.IsSpecified() {
		t.Fatal("non-blank optional text should be specified")
	}
	if text.String() != "note" {
		t.Fatalf("optional text = %q, want %q", text.String(), "note")
	}
}
