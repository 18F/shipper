package main

import (
	"github.com/codegangsta/cli"
	"os"
)

func Setup(context *cli.Context) {
	directories := [...]string{"releases", "shared"}

	config := LoadConfig(context)

	for _, dir := range directories {
		os.MkdirAll(config.AppPath+"/"+dir, os.FileMode(0755))
	}
}
