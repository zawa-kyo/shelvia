package domain

import "testing"

func TestNewAllowedValues(t *testing.T) {
	t.Run("設定値を正規化して検索できる", func(t *testing.T) {
		got, err := NewAllowedValues(AllowedValuesInput{
			Genres:     []string{"  Novel  "},
			Publishers: []string{"  Example Publisher  "},
			Imprints:   []string{"  Example Paperback  "},
			Editions: []EditionInput{
				{Name: "  Paperback  ", ImprintRequired: true},
				{Name: "  Hardcover  ", ImprintRequired: false},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		genre, err := NewGenre("Novel")
		if err != nil {
			t.Fatalf("NewGenre: %v", err)
		}
		if !got.ContainsGenre(genre) {
			t.Fatal("genre should be allowed")
		}

		publisher, err := NewPublisher("Example Publisher")
		if err != nil {
			t.Fatalf("NewPublisher: %v", err)
		}
		if !got.ContainsPublisher(publisher) {
			t.Fatal("publisher should be allowed")
		}

		imprint, err := NewImprint("Example Paperback")
		if err != nil {
			t.Fatalf("NewImprint: %v", err)
		}
		if !got.ContainsImprint(imprint) {
			t.Fatal("imprint should be allowed")
		}

		paperback, err := NewEdition("Paperback")
		if err != nil {
			t.Fatalf("NewEdition: %v", err)
		}
		paperbackRule, ok := got.FindEdition(paperback)
		if !ok {
			t.Fatal("paperback should be allowed")
		}
		if paperbackRule.Name().String() != "Paperback" {
			t.Fatalf("edition = %q, want %q", paperbackRule.Name().String(), "Paperback")
		}
		if !paperbackRule.ImprintRequired() {
			t.Fatal("paperback should require imprint")
		}

		hardcover, err := NewEdition("Hardcover")
		if err != nil {
			t.Fatalf("NewEdition: %v", err)
		}
		hardcoverRule, ok := got.FindEdition(hardcover)
		if !ok {
			t.Fatal("hardcover should be allowed")
		}
		if hardcoverRule.ImprintRequired() {
			t.Fatal("hardcover should not require imprint")
		}
	})
}

func TestNewAllowedValuesRejectsInvalidConfigValues(t *testing.T) {
	tests := []struct {
		name  string
		input AllowedValuesInput
	}{
		{
			name: "空文字のジャンルは拒否する",
			input: AllowedValuesInput{
				Genres: []string{"Novel", " "},
			},
		},
		{
			name: "重複した出版社は拒否する",
			input: AllowedValuesInput{
				Publishers: []string{"Example Publisher", "Example Publisher"},
			},
		},
		{
			name: "空文字の版は拒否する",
			input: AllowedValuesInput{
				Editions: []EditionInput{{Name: " "}},
			},
		},
		{
			name: "重複したジャンルは拒否する",
			input: AllowedValuesInput{
				Genres: []string{"Novel", " Novel "},
			},
		},
		{
			name: "空文字の出版社は拒否する",
			input: AllowedValuesInput{
				Publishers: []string{" "},
			},
		},
		{
			name: "空文字の印刷所は拒否する",
			input: AllowedValuesInput{
				Imprints: []string{" "},
			},
		},
		{
			name: "重複した印刷所は拒否する",
			input: AllowedValuesInput{
				Imprints: []string{"Example Paperback", " Example Paperback "},
			},
		},
		{
			name: "重複した版は拒否する",
			input: AllowedValuesInput{
				Editions: []EditionInput{{Name: "Paperback"}, {Name: " Paperback "}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewAllowedValues(tt.input); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
