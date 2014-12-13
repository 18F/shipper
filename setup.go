package main

import (
	"os"
)

func Setup(config *Config) {
	directories := [...]string{"releases", "shared"}

	for _, dir := range directories {
		os.MkdirAll(config.AppPath+"/"+dir, os.FileMode(0755))
	}
}
