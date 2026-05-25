package localfs

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

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

func tomlEscape(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	return strings.ReplaceAll(value, `"`, `\"`)
}
