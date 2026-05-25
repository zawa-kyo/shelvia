package domain

import "testing"

func TestAllowedValues(t *testing.T) {
	t.Run("前後の空白は候補名に含めない", func(t *testing.T) {
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
			t.Fatalf("候補値を読み込めませんでした: %v", err)
		}

		genre, err := NewGenre("Novel")
		if err != nil {
			t.Fatalf("ジャンル名を準備できませんでした: %v", err)
		}
		if !got.ContainsGenre(genre) {
			t.Fatal("ジャンル候補として扱われていません")
		}

		publisher, err := NewPublisher("Example Publisher")
		if err != nil {
			t.Fatalf("出版社名を準備できませんでした: %v", err)
		}
		if !got.ContainsPublisher(publisher) {
			t.Fatal("出版社候補として扱われていません")
		}

		imprint, err := NewImprint("Example Paperback")
		if err != nil {
			t.Fatalf("レーベル名を準備できませんでした: %v", err)
		}
		if !got.ContainsImprint(imprint) {
			t.Fatal("レーベル候補として扱われていません")
		}

		paperback, err := NewEdition("Paperback")
		if err != nil {
			t.Fatalf("判型を準備できませんでした: %v", err)
		}
		paperbackRule, ok := got.FindEdition(paperback)
		if !ok {
			t.Fatal("判型候補として扱われていません")
		}
		if paperbackRule.Name().String() != "Paperback" {
			t.Fatalf("判型 = %q, want %q", paperbackRule.Name().String(), "Paperback")
		}
		if !paperbackRule.ImprintRequired() {
			t.Fatal("この判型ではレーベルが必須です")
		}

		hardcover, err := NewEdition("Hardcover")
		if err != nil {
			t.Fatalf("判型を準備できませんでした: %v", err)
		}
		hardcoverRule, ok := got.FindEdition(hardcover)
		if !ok {
			t.Fatal("判型候補として扱われていません")
		}
		if hardcoverRule.ImprintRequired() {
			t.Fatal("この判型ではレーベルを必須にしない想定です")
		}
	})

	t.Run("まだ使わない候補欄は空でもよい", func(t *testing.T) {
		if _, err := NewAllowedValues(AllowedValuesInput{}); err != nil {
			t.Fatalf("空の候補欄を読み込めませんでした: %v", err)
		}
	})
}

func TestInvalidAllowedValues(t *testing.T) {
	tests := []struct {
		name  string
		input AllowedValuesInput
	}{
		{
			name: "空のジャンル名は候補にできない",
			input: AllowedValuesInput{
				Genres: []string{"Novel", " "},
			},
		},
		{
			name: "同じ出版社名は候補に重複して書けない",
			input: AllowedValuesInput{
				Publishers: []string{"Example Publisher", "Example Publisher"},
			},
		},
		{
			name: "空の判型名は候補にできない",
			input: AllowedValuesInput{
				Editions: []EditionInput{{Name: " "}},
			},
		},
		{
			name: "空白違いだけのジャンル名は同じ候補として扱う",
			input: AllowedValuesInput{
				Genres: []string{"Novel", " Novel "},
			},
		},
		{
			name: "空の出版社名は候補にできない",
			input: AllowedValuesInput{
				Publishers: []string{" "},
			},
		},
		{
			name: "空のレーベル名は候補にできない",
			input: AllowedValuesInput{
				Imprints: []string{" "},
			},
		},
		{
			name: "空白違いだけのレーベル名は同じ候補として扱う",
			input: AllowedValuesInput{
				Imprints: []string{"Example Paperback", " Example Paperback "},
			},
		},
		{
			name: "空白違いだけの判型名は同じ候補として扱う",
			input: AllowedValuesInput{
				Editions: []EditionInput{{Name: "Paperback"}, {Name: " Paperback "}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewAllowedValues(tt.input); err == nil {
				t.Fatal("候補として受け入れられてしまいました")
			}
		})
	}
}
