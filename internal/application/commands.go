package application

import "time"

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
)

// Command contains the validated command-line intent.
type Command struct {
	Kind      CommandKind
	ShelfRoot string
	Title     string
	ReadDate  time.Time
	Where     string
}

// Table is a presentation-neutral tabular result.
type Table struct {
	Headers []string
	Rows    [][]string
}

// Output is the presentation-neutral command result.
type Output struct {
	Message string
	Table   Table
}
