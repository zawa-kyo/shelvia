package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/zawa-kyo/shelvia/internal/application"
)

const shelfEnv = "SHELVIA_DIR"

// Parses CLI arguments into application commands.
func Parse(args []string) (application.Command, error) {
	flags := flag.NewFlagSet("shelvia", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	shelf := flags.String("shelf", "", "shelf root")
	if err := flags.Parse(args); err != nil {
		return application.Command{}, err
	}

	rest := flags.Args()
	if len(rest) == 0 {
		return application.Command{}, fmt.Errorf("missing command")
	}

	switch rest[0] {
	case "init":
		if len(rest) > 2 {
			return application.Command{}, fmt.Errorf("usage: shelvia init [PATH]")
		}
		root := "."
		if len(rest) == 2 {
			root = rest[1]
		}
		return application.Command{Kind: application.Init, ShelfRoot: root}, nil
	case "add":
		command, err := parseAdd(rest[1:])
		if err != nil {
			return application.Command{}, err
		}
		command.ShelfRoot = resolveRoot(*shelf)
		return requireRoot(command)
	case "validate":
		if len(rest) != 1 {
			return application.Command{}, fmt.Errorf("usage: shelvia validate")
		}
		return requireRoot(application.Command{Kind: application.Validate, ShelfRoot: resolveRoot(*shelf)})
	case "list":
		if len(rest) != 1 {
			return application.Command{}, fmt.Errorf("usage: shelvia list")
		}
		return requireRoot(application.Command{Kind: application.List, ShelfRoot: resolveRoot(*shelf)})
	case "search":
		if len(rest) != 2 {
			return application.Command{}, fmt.Errorf("usage: shelvia search SQL_FRAGMENT")
		}
		return requireRoot(application.Command{Kind: application.Search, ShelfRoot: resolveRoot(*shelf), Where: rest[1]})
	case "show":
		if len(rest) != 2 {
			return application.Command{}, fmt.Errorf("usage: shelvia show TITLE")
		}
		return requireRoot(application.Command{Kind: application.Show, ShelfRoot: resolveRoot(*shelf), Title: rest[1]})
	case "path":
		if len(rest) != 2 {
			return application.Command{}, fmt.Errorf("usage: shelvia path TITLE")
		}
		return requireRoot(application.Command{Kind: application.Path, ShelfRoot: resolveRoot(*shelf), Title: rest[1]})
	default:
		return application.Command{}, fmt.Errorf("unknown command: %s", rest[0])
	}
}

func parseAdd(args []string) (application.Command, error) {
	var title string
	var readDate string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--date":
			if i+1 >= len(args) {
				return application.Command{}, fmt.Errorf("add requires --date")
			}
			readDate = args[i+1]
			i++
		default:
			if strings.HasPrefix(args[i], "-") {
				return application.Command{}, fmt.Errorf("unknown add option: %s", args[i])
			}
			if title != "" {
				return application.Command{}, fmt.Errorf("usage: shelvia add TITLE --date YYYY-MM-DD")
			}
			title = args[i]
		}
	}
	if title == "" {
		return application.Command{}, fmt.Errorf("usage: shelvia add TITLE --date YYYY-MM-DD")
	}
	if readDate == "" {
		return application.Command{}, fmt.Errorf("add requires --date")
	}
	date, err := time.Parse("2006-01-02", readDate)
	if err != nil {
		return application.Command{}, fmt.Errorf("invalid --date: %s", readDate)
	}
	return application.Command{Kind: application.Add, Title: title, ReadDate: date}, nil
}

func resolveRoot(explicit string) string {
	if explicit != "" {
		return explicit
	}
	return os.Getenv(shelfEnv)
}

func requireRoot(command application.Command) (application.Command, error) {
	if command.ShelfRoot == "" {
		return application.Command{}, fmt.Errorf("shelf root is not set; pass --shelf or set %s", shelfEnv)
	}
	return command, nil
}
