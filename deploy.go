package main

import (
	"bytes"
	"github.com/codegangsta/cli"
	"github.com/dlapiduz/go-github/github"
	"log"
	"os/exec"
	"strings"
)

func findRev(config *Config) string {
	log.Println("Finding last revision")

	if config.UseGithubAPI {
		client := config.GetGithubClient()

		user, repo := config.ParseGithubInfo()
		// Get the latest deployment for the environment
		dep_list := github.DeploymentsListOptions{
			Environment: config.Environment,
		}
		deployments, _, err := client.Repositories.ListDeployments(user, repo, &dep_list)
		PanicOn(err)

		// Do we have a deployment?
		if len(deployments) > 0 {

			log.Println("Deployment found!")

			// Check the statuses of the last one
			statuses, _, err := client.Repositories.ListDeploymentStatuses(user, repo, *deployments[0].ID, nil)
			PanicOn(err)

			// Are there statuses?
			if len(statuses) > 0 {
				log.Println("There are statuses!")
				for _, status := range statuses {
					log.Println(status)
				}
			} else {
				log.Println("No status found...")
				// Lets deploy but before lets ping the api
				err := createDeployStatus(config, &deployments[0], "pending")
				PanicOn(err)
				log.Println("Set to pending")

				// Actually deploy
				// Lets mark the deploy as a success
				err = createDeployStatus(config, &deployments[0], "success")
				PanicOn(err)
				log.Println("Set to success")

			}

			return ""
		} else {
			return ""
		}

	} else {
		args := []string{
			"ls-remote",
			config.GitUrl,
			config.Revision,
		}

		rev_ls, err := exec.Command("git", args...).Output()
		PanicOn(err)
		revision := strings.Split(string(rev_ls), "\t")[0]

		return revision

	}
}

func createDeployStatus(config *Config, deployment *github.Deployment, status string) error {
	client := config.GetGithubClient()
	user, repo := config.ParseGithubInfo()

	req := github.DeploymentStatusRequest{
		State:       &status,
		Description: github.String("{\"server\": \"" + config.ServerId + "\"}"),
	}
	_, _, err := client.Repositories.CreateDeploymentStatus(user, repo, *deployment.ID, &req)
	return err
}

func doCheckout(config *Config, checkout_path string) error {
	log.Println("Checking out code")
	args := []string{
		"clone",
		"--depth=5",
		config.GitUrl,
		checkout_path,
	}

	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		log.Println(string(stderr.Bytes()))
		log.Fatal(err)
	}

	log.Println("Deployed!")
	return err
}

func Deploy(context *cli.Context) {
	config := LoadConfig(context)
	revision := findRev(&config)

	_ = revision
	// checkout_path := config.AppPath + "/releases/" + revision

	// if _, err := os.Stat(checkout_path); err != nil {
	// 	if os.IsNotExist(err) {
	// 		doCheckout(&config, checkout_path)
	// 	}
	// } else {
	// 	log.Println("Revision already exists")
	// }

}
