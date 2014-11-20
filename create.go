package main

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/google/go-github/github"
)

func Create(context *cli.Context) {
	config, err := LoadConfig(context)
	if err != nil {
		log.Println("There was an error loading the config")
		log.Println(err)
		return
	}
	var environment string

	if context.String("ref") == "" {
		log.Println("Ref is required, exiting")
		return
	}

	if context.String("environment") != "" {
		environment = context.String("environment")
	} else {
		environment = config.Environment
	}

	client := config.GetGithubClient()
	user, repo := config.ParseGithubInfo()

	log.Println("Creating deploy")

	status_req := github.DeploymentRequest{
		Ref:         github.String(context.String("ref")),
		Task:        github.String("deploy"),
		AutoMerge:   github.Bool(false),
		Environment: &environment,
	}

	deployment, _, err := client.Repositories.CreateDeployment(user, repo, &status_req)

	log.Println("Deploy created")
	log.Println(deployment)
	log.Println(err)
}
