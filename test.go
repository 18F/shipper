package main

import (
	"github.com/codegangsta/cli"
	"github.com/dlapiduz/go-github/github"
	"log"
)

func Test(context *cli.Context) {
	config := LoadConfig(context)

	client := config.GetGithubClient()
	user, repo := config.ParseGithubInfo()

	log.Println("Creating deploy")

	status_req := github.DeploymentRequest{
		Ref:         github.String("23458be"),
		Task:        github.String("deploy"),
		AutoMerge:   github.Bool(false),
		Environment: &config.Environment,
	}

	deployment, _, err := client.Repositories.CreateDeployment(user, repo, &status_req)

	log.Println("Deploy created")
	log.Println(deployment)
	log.Println(err)
}
