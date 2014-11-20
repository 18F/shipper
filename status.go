package main

import (
	"encoding/json"

	"github.com/google/go-github/github"
)

type StatusDescription struct {
	Server string
}

func createDeployStatus(config *Config, deployment *github.Deployment, status string) {
	client := config.GetGithubClient()
	user, repo := config.ParseGithubInfo()

	desc, _ := json.Marshal(StatusDescription{Server: config.ServerId})

	req := github.DeploymentStatusRequest{
		State:       &status,
		Description: github.String(string(desc)),
	}
	_, _, err := client.Repositories.CreateDeploymentStatus(user, repo, *deployment.ID, &req)
	PanicOn(err)
}

// findNewDeployment uses the github deployments api to look for a new deployment
// that has no status for the server that this script is running on.
func findNewDeployment(config *Config) (*github.Deployment, error) {
	client := config.GetGithubClient()
	user, repo := config.ParseGithubInfo()

	// Get the latest deployment for the environment
	dep_list := github.DeploymentsListOptions{
		Environment: config.Environment,
	}
	deployments, _, err := client.Repositories.ListDeployments(user, repo, &dep_list)
	if err != nil {
		// something went wrong getting the deployments
		return nil, err
	}

	// Do we have a deployment?
	if len(deployments) > 0 {
		// Get latest deployment
		deployment := &deployments[0]

		status := getLatestStatus(config, deployment)

		if status != nil {
			// there is some status for this deployment on this server so
			// no need to do anything
			// TODO: smarter handling of error or stuck pending deploys
			if *status.State == "pending" {
				return deployment, nil
			}
			return nil, nil
		} else {
			// there is no status, lets deploy
			return deployment, nil
		}
	} else {
		// No deployment for this environment
		return nil, nil
	}

}

func getLatestStatus(config *Config, deployment *github.Deployment) *github.DeploymentStatus {
	client := config.GetGithubClient()
	user, repo := config.ParseGithubInfo()

	// find all statuses
	statuses, _, err := client.Repositories.ListDeploymentStatuses(user, repo, *deployment.ID, nil)
	PanicOn(err)
	// Are there statuses?
	if len(statuses) > 0 {
		// find the latest status for this server
		for _, status := range statuses {
			var desc StatusDescription
			json.Unmarshal([]byte(*status.Description), &desc)
			if desc.Server == config.ServerId {
				return &status
			}
		}
	}

	// No relevant status found
	return nil
}
