package main

import (
	"github.com/codegangsta/cli"
	"log"
	"time"
)

func Run(context *cli.Context) {
	config, _ := LoadConfig(context)
	log.Println("Running shipper with an interval of ", config.Interval, " minutes")
	c := time.Tick(time.Duration(config.Interval) * time.Minute)
	for _ = range c {
		Deploy(&config)
	}
}
