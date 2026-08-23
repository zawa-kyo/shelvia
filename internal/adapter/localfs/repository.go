package localfs

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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
	configPath := filepath.Join(root, configFileName)
	examplePath := filepath.Join(root, "example.toml")
	for _, path := range []string{configPath, examplePath} {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("%s already exists", filepath.ToSlash(path))
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	if err := writeNewFile(configPath, []byte(configTemplate)); err != nil {
		return err
	}
	return writeNewFile(examplePath, []byte(exampleTemplate))
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
	config, err := readConfig(root, filepath.Join(root, configFileName))
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
			return fmt.Errorf("%s: misplaced config.toml", relativePath(root, path))
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
