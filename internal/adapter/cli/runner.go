package cli

import (
	"io"

	"github.com/zawa-kyo/shelvia/internal/adapter/presentation"
	"github.com/zawa-kyo/shelvia/internal/application"
)

type commandRunner interface {
	Run(application.Command) (application.Output, error)
}

// Parses arguments, runs the application service, and writes output.
func Run(args []string, service commandRunner, out io.Writer, errOut io.Writer) int {
	command, err := Parse(args)
	if err != nil {
		presentation.WriteError(errOut, err)
		return 1
	}

	output, err := service.Run(command)
	if err != nil {
		presentation.WriteError(errOut, err)
		return 1
	}

	if command.Format == "csv" {
		err = presentation.WriteCSV(out, output)
	} else {
		err = presentation.Write(out, output)
	}
	if err != nil {
		presentation.WriteError(errOut, err)
		return 1
	}

	return 0
}
