package localfs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/pelletier/go-toml/v2"
	"github.com/zawa-kyo/shelvia/internal/application"
)

const configFileName = "config.toml"

// Repository reads and writes shelves on the local filesystem.
type Repository struct{}

// Creates config.toml and example.toml without overwriting existing files.
func (Repository) Init(root string) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	if err := writeNewFile(filepath.Join(root, configFileName), []byte(configTemplate)); err != nil {
		return err
	}
	return writeNewFile(filepath.Join(root, "example.toml"), []byte(exampleTemplate))
}

// Creates a TOML template for a single book.
func (Repository) CreateBook(root, title string, readDate time.Time) (string, error) {
	filename, err := safeBookFileName(title)
	if err != nil {
		return "", err
	}
	path := filepath.Join(root, filename)
	content := fmt.Sprintf(bookTemplate, tomlEscape(title), readDate.Format("2006-01-02"))
	if err := writeNewFile(path, []byte(content)); err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filename, nil
	}
	return filepath.ToSlash(rel), nil
}

// Loads config.toml and every book TOML under the shelf root.
func (Repository) Load(root string) (application.ShelfData, error) {
	config, err := readConfig(filepath.Join(root, configFileName))
	if err != nil {
		return application.ShelfData{}, err
	}

	paths, err := findBookFiles(root)
	if err != nil {
		return application.ShelfData{}, err
	}
	books := make([]application.BookData, 0, len(paths))
	for _, path := range paths {
		book, err := readBook(root, path)
		if err != nil {
			return application.ShelfData{}, err
		}
		books = append(books, book)
	}
	return application.ShelfData{Config: config, Books: books}, nil
}

func readConfig(path string) (application.AllowedValuesData, error) {
	var raw configTOML
	if err := decodeTOML(path, &raw); err != nil {
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

func readBook(root, path string) (application.BookData, error) {
	var raw bookTOML
	if err := decodeTOML(path, &raw); err != nil {
		return application.BookData{}, err
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = path
	}
	return application.BookData{
		Path:       filepath.ToSlash(rel),
		Title:      raw.Title,
		Author:     raw.Author,
		Rating:     raw.Rating,
		ReadDate:   localDateTime(raw.ReadDate),
		Genre:      raw.Genre,
		Publisher:  raw.Publisher,
		Edition:    raw.Edition,
		Imprint:    raw.Imprint,
		Series:     raw.Series,
		Translator: raw.Translator,
		Summary:    raw.Thoughts.Summary,
		Body:       raw.Thoughts.Body,
	}, nil
}

func decodeTOML(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := toml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("%s: %w", filepath.ToSlash(path), err)
	}
	return nil
}

func findBookFiles(root string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(path), ".toml") {
			return nil
		}
		if filepath.Base(path) == configFileName {
			if samePath(filepath.Dir(path), root) {
				return nil
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			return fmt.Errorf("%s: misplaced config.toml", filepath.ToSlash(rel))
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

func writeNewFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("%s already exists", filepath.ToSlash(path))
		}
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}

func safeBookFileName(title string) (string, error) {
	name := strings.TrimSpace(title)
	if strings.EqualFold(filepath.Ext(name), ".toml") {
		name = strings.TrimSpace(strings.TrimSuffix(name, filepath.Ext(name)))
	}
	if name == "" || name == "." || name == ".." {
		return "", fmt.Errorf("invalid title for file name: %q", title)
	}
	if strings.HasSuffix(name, ".") {
		return "", fmt.Errorf("invalid trailing character in file name: %q", title)
	}
	if isWindowsReservedName(name) {
		return "", fmt.Errorf("reserved file name: %s", name)
	}
	for _, r := range name {
		if r == 0 || unicode.IsControl(r) || strings.ContainsRune(`<>:"/\|?*`, r) {
			return "", fmt.Errorf("invalid character in file name: %q", r)
		}
	}
	return name + ".toml", nil
}

func isWindowsReservedName(name string) bool {
	base := strings.ToUpper(strings.TrimSpace(name))
	base = strings.TrimSuffix(base, filepath.Ext(base))
	switch base {
	case "CON", "PRN", "AUX", "NUL":
		return true
	}
	for i := 1; i <= 9; i++ {
		if base == fmt.Sprintf("COM%d", i) || base == fmt.Sprintf("LPT%d", i) {
			return true
		}
	}
	return false
}

func localDateTime(date toml.LocalDate) time.Time {
	if date.Year == 0 || date.Month == 0 || date.Day == 0 {
		return time.Time{}
	}
	return date.AsTime(time.UTC)
}

func samePath(left, right string) bool {
	leftAbs, leftErr := filepath.Abs(left)
	rightAbs, rightErr := filepath.Abs(right)
	if leftErr != nil || rightErr != nil {
		return filepath.Clean(left) == filepath.Clean(right)
	}
	return filepath.Clean(leftAbs) == filepath.Clean(rightAbs)
}

func tomlEscape(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
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

type bookTOML struct {
	Title      string         `toml:"title"`
	Author     string         `toml:"author"`
	Rating     int            `toml:"rating"`
	ReadDate   toml.LocalDate `toml:"read_date"`
	Genre      string         `toml:"genre"`
	Publisher  string         `toml:"publisher"`
	Edition    string         `toml:"edition"`
	Imprint    string         `toml:"imprint"`
	Series     string         `toml:"series"`
	Translator string         `toml:"translator"`
	Thoughts   thoughtsTOML   `toml:"thoughts"`
}

type thoughtsTOML struct {
	Summary string `toml:"summary"`
	Body    string `toml:"body"`
}
