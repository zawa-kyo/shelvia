package localfs

import (
	"fmt"

	"github.com/zawa-kyo/shelvia/internal/application"
)

func readConfig(root, path string) (application.AllowedValuesData, error) {
	var raw configTOML
	if err := decodeTOML(root, path, &raw); err != nil {
		return application.AllowedValuesData{}, err
	}
	if raw.Kind != "shelvia-config" {
		return application.AllowedValuesData{}, fmt.Errorf("config.toml: kind must be \"shelvia-config\"")
	}
	editions := make([]application.EditionData, len(raw.Values.Editions))
	for i, edition := range raw.Values.Editions {
		editions[i] = application.EditionData{
			Name:            edition.Name,
			ImprintRequired: edition.ImprintRequired,
		}
	}
	return application.AllowedValuesData{
		Genres:     raw.Values.Genres,
		Publishers: raw.Values.Publishers,
		Imprints:   raw.Values.Imprints,
		Editions:   editions,
	}, nil
}

type configTOML struct {
	Kind   string           `toml:"kind"`
	Values configValuesTOML `toml:"values"`
}

type configValuesTOML struct {
	Genres     []string      `toml:"genres"`
	Publishers []string      `toml:"publishers"`
	Imprints   []string      `toml:"imprints"`
	Editions   []editionTOML `toml:"editions"`
}

type editionTOML struct {
	Name            string `toml:"name"`
	ImprintRequired bool   `toml:"imprint_required"`
}
