package main

import (
	"github.com/codegangsta/cli"
	"github.com/dlapiduz/go-github/github"
	"log"
)

func Create(context *cli.Context) {
	config := LoadConfig(context)
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
