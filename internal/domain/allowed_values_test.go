package domain

import "testing"

func TestNewAllowedValuesRejectsInvalidConfigValues(t *testing.T) {
	tests := []struct {
		name  string
		input AllowedValuesInput
	}{
		{
			name: "blank genre",
			input: AllowedValuesInput{
				Genres: []string{"Novel", " "},
			},
		},
		{
			name: "duplicate publisher",
			input: AllowedValuesInput{
				Publishers: []string{"Example Publisher", "Example Publisher"},
			},
		},
		{
			name: "blank edition",
			input: AllowedValuesInput{
				Editions: []EditionInput{{Name: " "}},
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
