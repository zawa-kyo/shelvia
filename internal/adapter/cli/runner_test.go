package cli

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zawa-kyo/shelvia/internal/application"
)

func TestRun(t *testing.T) {
	t.Run("解析したコマンドを実行して結果を出力する", func(t *testing.T) {
		service := &stubRunner{output: application.Output{Message: "Done."}}
		var stdout, stderr bytes.Buffer

		code := Run([]string{"--shelf", "my-shelf", "validate"}, service, &stdout, &stderr)

		require.Equal(t, 0, code)
		require.Equal(t, application.Command{Kind: application.Validate, ShelfRoot: "my-shelf"}, service.command)
		require.Equal(t, 1, service.calls)
		require.Equal(t, "Done.\n", stdout.String())
		require.Empty(t, stderr.String())
	})

	t.Run("引数が不正な場合は実行せずエラーを出力する", func(t *testing.T) {
		service := &stubRunner{}
		var stdout, stderr bytes.Buffer

		code := Run([]string{"unknown"}, service, &stdout, &stderr)

		require.Equal(t, 1, code)
		require.Zero(t, service.calls)
		require.Empty(t, stdout.String())
		require.Equal(t, "unknown command: unknown\n", stderr.String())
	})

	t.Run("実行エラーを標準エラーに出力する", func(t *testing.T) {
		service := &stubRunner{err: errors.New("load failed")}
		var stdout, stderr bytes.Buffer

		code := Run([]string{"--shelf", "my-shelf", "validate"}, service, &stdout, &stderr)

		require.Equal(t, 1, code)
		require.Equal(t, 1, service.calls)
		require.Empty(t, stdout.String())
		require.Equal(t, "load failed\n", stderr.String())
	})

	t.Run("CSV出力の書き込みエラーを標準エラーに出力する", func(t *testing.T) {
		service := &stubRunner{output: application.Output{Kind: application.List}}
		var stderr bytes.Buffer

		code := Run([]string{"--shelf", "my-shelf", "list", "--format", "csv"}, service, failingWriter{}, &stderr)

		require.Equal(t, 1, code)
		require.Equal(t, "csv", service.command.Format)
		require.Equal(t, "write failed\n", stderr.String())
	})
}

type stubRunner struct {
	command application.Command
	output  application.Output
	err     error
	calls   int
}

func (runner *stubRunner) Run(command application.Command) (application.Output, error) {
	runner.command = command
	runner.calls++
	return runner.output, runner.err
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
