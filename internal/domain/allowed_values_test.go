package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
		require.NoError(t, err, "候補値を読み込めませんでした")

		genre, err := NewGenre("Novel")
		require.NoError(t, err, "ジャンル名を準備できませんでした")
		require.True(t, got.ContainsGenre(genre), "ジャンル候補として扱われていません")

		publisher, err := NewPublisher("Example Publisher")
		require.NoError(t, err, "出版社名を準備できませんでした")
		require.True(t, got.ContainsPublisher(publisher), "出版社候補として扱われていません")

		imprint, err := NewImprint("Example Paperback")
		require.NoError(t, err, "レーベル名を準備できませんでした")
		require.True(t, got.ContainsImprint(imprint), "レーベル候補として扱われていません")

		paperback, err := NewEdition("Paperback")
		require.NoError(t, err, "判型を準備できませんでした")

		paperbackRule, ok := got.FindEdition(paperback)
		require.True(t, ok, "判型候補として扱われていません")
		require.Equal(t, "Paperback", paperbackRule.Name().String())
		require.True(t, paperbackRule.ImprintRequired(), "この判型ではレーベルが必須です")

		hardcover, err := NewEdition("Hardcover")
		require.NoError(t, err, "判型を準備できませんでした")

		hardcoverRule, ok := got.FindEdition(hardcover)
		require.True(t, ok, "判型候補として扱われていません")
		require.False(t, hardcoverRule.ImprintRequired(), "この判型ではレーベルを必須にしない想定です")
	})

	t.Run("まだ使わない候補欄は空でもよい", func(t *testing.T) {
		_, err := NewAllowedValues(AllowedValuesInput{})
		require.NoError(t, err, "空の候補欄を読み込めませんでした")
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
			_, err := NewAllowedValues(tt.input)
			require.Error(t, err, "候補として受け入れられてしまいました")
		})
	}
}
