package application

import (
	"time"

	"github.com/zawa-kyo/shelvia/internal/domain"
)

// CommandKind identifies a supported CLI command.
type CommandKind int

const (
	Init CommandKind = iota
	Add
	Validate
	List
	Search
	Show
	Path
	Export
)

// Command contains the validated command-line intent.
type Command struct {
	Kind      CommandKind
	ShelfRoot string
	Title     string
	ReadDate  time.Time
	Where     string
	Format    string
}

// Output contains the data produced by a command.
type Output struct {
	Kind    CommandKind
	Message string
	Books   []domain.Book
	Book    domain.Book
}
