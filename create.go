package main

import (
	"log"

	"github.com/codegangsta/cli"
)

func Create(context *cli.Context, config *Config) {
	ref := context.String("ref")
	if ref == "" {
		log.Println("Ref is required, exiting")
		return
	}

	var env string
	if context.String("environment") != "" {
		env = context.String("environment")
	} else {
		env = config.Environment
	}

	log.Println("Creating deploy")

	err := config.Backend.CreateDeployment(ref, env)

	if err != nil {
		log.Println("There was an error creating the deploy")
		log.Println(err)
		return
	}

	log.Println("Deploy Created")
}
