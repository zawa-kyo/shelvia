package main

import (
	"os"

	"github.com/zawa-kyo/shelvia/internal/adapter/cli"
	"github.com/zawa-kyo/shelvia/internal/adapter/localfs"
	"github.com/zawa-kyo/shelvia/internal/adapter/queryengine"
	"github.com/zawa-kyo/shelvia/internal/application"
)

func main() {
	service := application.NewService(localfs.Repository{}, queryengine.Store{})
	os.Exit(cli.Run(os.Args[1:], service, os.Stdout, os.Stderr))
}
