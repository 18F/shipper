package main

import (
	"github.com/codegangsta/cli"
	"github.com/dlapiduz/go-github/github"
	"log"
)

func NewDeploy(context *cli.Context) {
	config := LoadConfig(context)
	var ref, environment string

	if context.String("ref") != "" {
		ref = context.String("ref")
	} else {
		ref = config.Revision
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
		Ref:         &ref,
		Task:        github.String("deploy"),
		AutoMerge:   github.Bool(false),
		Environment: &environment,
	}

	deployment, _, err := client.Repositories.CreateDeployment(user, repo, &status_req)

	log.Println("Deploy created")
	log.Println(deployment)
	log.Println(err)
}
