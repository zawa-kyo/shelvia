package domain

import "fmt"

// Raw input for an edition controlled value.
type EditionInput struct {
	Name            string
	ImprintRequired bool
}

// Raw input for config-defined controlled values.
type AllowedValuesInput struct {
	Genres     []string
	Publishers []string
	Imprints   []string
	Editions   []EditionInput
}

// Describes an allowed edition and whether it requires an imprint.
type EditionRule struct {
	name            Edition
	imprintRequired bool
}

// Returns the controlled edition value.
func (rule EditionRule) Name() Edition {
	return rule.name
}

// Reports whether books with this edition must specify an imprint.
func (rule EditionRule) ImprintRequired() bool {
	return rule.imprintRequired
}

// Controlled values accepted by the domain model.
type AllowedValues struct {
	genres     map[string]struct{}
	publishers map[string]struct{}
	imprints   map[string]struct{}
	editions   map[string]EditionRule
}

// Creates a controlled-value set and rejects blank or duplicate entries.
func NewAllowedValues(input AllowedValuesInput) (AllowedValues, error) {
	genres, err := buildRequiredStringSet("genre", input.Genres, NewGenre)
	if err != nil {
		return AllowedValues{}, err
	}
	publishers, err := buildRequiredStringSet("publisher", input.Publishers, NewPublisher)
	if err != nil {
		return AllowedValues{}, err
	}
	imprints, err := buildRequiredStringSet("imprint", input.Imprints, func(value string) (stringer, error) {
		imprint, err := NewImprint(value)
		if err != nil {
			return nil, err
		}
		if !imprint.IsSpecified() {
			return nil, validationError("imprint", "must not be blank")
		}
		return imprint, nil
	})
	if err != nil {
		return AllowedValues{}, err
	}

	editions := make(map[string]EditionRule, len(input.Editions))
	for _, item := range input.Editions {
		edition, err := NewEdition(item.Name)
		if err != nil {
			return AllowedValues{}, err
		}
		if !edition.IsSpecified() {
			return AllowedValues{}, validationError("edition", "must not be blank")
		}
		name := edition.String()
		if _, exists := editions[name]; exists {
			return AllowedValues{}, validationError("edition", fmt.Sprintf("duplicate value: %s", name))
		}
		editions[name] = EditionRule{name: edition, imprintRequired: item.ImprintRequired}
	}

	return AllowedValues{
		genres:     genres,
		publishers: publishers,
		imprints:   imprints,
		editions:   editions,
	}, nil
}

// Reports whether the genre is configured as an allowed value.
func (values AllowedValues) ContainsGenre(genre Genre) bool {
	_, ok := values.genres[genre.String()]
	return ok
}

// Reports whether the publisher is configured as an allowed value.
func (values AllowedValues) ContainsPublisher(publisher Publisher) bool {
	_, ok := values.publishers[publisher.String()]
	return ok
}

// Reports whether the imprint is configured as an allowed value.
func (values AllowedValues) ContainsImprint(imprint Imprint) bool {
	_, ok := values.imprints[imprint.String()]
	return ok
}

// Returns the configured rule for the edition.
func (values AllowedValues) FindEdition(edition Edition) (EditionRule, bool) {
	rule, ok := values.editions[edition.String()]
	return rule, ok
}

type stringer interface {
	String() string
}

func buildRequiredStringSet[T stringer](field string, rawValues []string, newValue func(string) (T, error)) (map[string]struct{}, error) {
	values := make(map[string]struct{}, len(rawValues))
	for _, raw := range rawValues {
		value, err := newValue(raw)
		if err != nil {
			return nil, err
		}
		key := value.String()
		if key == "" {
			return nil, validationError(field, "must not be blank")
		}
		if _, exists := values[key]; exists {
			return nil, validationError(field, fmt.Sprintf("duplicate value: %s", key))
		}
		values[key] = struct{}{}
	}
	return values, nil
}
