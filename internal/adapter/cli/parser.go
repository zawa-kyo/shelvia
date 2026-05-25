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
		if len(rest) != 2 {
			return application.Command{}, fmt.Errorf("usage: shelvia init PATH")
		}
		return application.Command{Kind: application.Init, ShelfRoot: rest[1]}, nil
	case "new":
		command, err := parseNew(rest[1:])
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
	case "query":
		command, err := parseQuery(rest[1:])
		if err != nil {
			return application.Command{}, err
		}
		command.ShelfRoot = resolveRoot(*shelf)
		return requireRoot(command)
	default:
		return application.Command{}, fmt.Errorf("unknown command: %s", rest[0])
	}
}

func parseNew(args []string) (application.Command, error) {
	var title string
	var readDate string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--read-date":
			if i+1 >= len(args) {
				return application.Command{}, fmt.Errorf("new requires --read-date")
			}
			readDate = args[i+1]
			i++
		default:
			if strings.HasPrefix(args[i], "-") {
				return application.Command{}, fmt.Errorf("unknown new option: %s", args[i])
			}
			if title != "" {
				return application.Command{}, fmt.Errorf("usage: shelvia new TITLE --read-date YYYY-MM-DD")
			}
			title = args[i]
		}
	}
	if title == "" {
		return application.Command{}, fmt.Errorf("usage: shelvia new TITLE --read-date YYYY-MM-DD")
	}
	if readDate == "" {
		return application.Command{}, fmt.Errorf("new requires --read-date")
	}
	date, err := time.Parse("2006-01-02", readDate)
	if err != nil {
		return application.Command{}, fmt.Errorf("invalid --read-date: %s", readDate)
	}
	return application.Command{Kind: application.New, Title: title, ReadDate: date}, nil
}

func parseQuery(args []string) (application.Command, error) {
	flags := flag.NewFlagSet("query", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	where := flags.String("where", "", "SQL-like where clause")
	if err := flags.Parse(args); err != nil {
		return application.Command{}, err
	}
	if *where == "" {
		return application.Command{}, fmt.Errorf("query requires --where")
	}
	if len(flags.Args()) != 0 {
		return application.Command{}, fmt.Errorf("usage: shelvia query --where SQL_FRAGMENT")
	}
	return application.Command{Kind: application.Query, Where: *where}, nil
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
