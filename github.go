package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"code.google.com/p/goauth2/oauth"
	"github.com/dlapiduz/httpcache"
	"github.com/google/go-github/github"
)

// Github Based Deployment Backend
// Uses the github deployments API to check for new deployments
type GithubBackend struct {
	Config *Config
	client *github.Client
}

// Internal method to get (and cache) the github client
func (b *GithubBackend) getClient() *github.Client {
	if b.client != nil {
		return b.client
	}
	gh_key := os.Getenv("GH_KEY")

	authTransport := &oauth.Transport{
		Token: &oauth.Token{AccessToken: gh_key},
	}

	memoryCacheTransport := httpcache.NewMemoryCacheTransport()
	memoryCacheTransport.Transport = authTransport

	httpClient := &http.Client{Transport: memoryCacheTransport}

	b.client = github.NewClient(httpClient)

	return b.client
}

// Find new deployments
// Uses the "Environment" setting from the config to query github
// and check if there are any deployments without any sttatus
// created by this server using the "ServerId" setting.
func (b *GithubBackend) FindNewDeployment() (*Deployment, error) {
	client := b.getClient()
	user, repo := b.parseGithubInfo()

	// Get the latest deployment for the environment
	dep_list := github.DeploymentsListOptions{
		Environment: b.Config.Environment,
	}

	deployments, _, err := client.Repositories.ListDeployments(user, repo, &dep_list)
	if err != nil {
		// something went wrong getting the deployments
		return nil, err
	}

	// Do we have a deployment?
	if len(deployments) > 0 {
		// Get latest deployment
		gh_deployment := &deployments[0]

		status, err := b.getLatestStatus(gh_deployment)
		if err != nil {
			// something went wrong getting the statuses
			return nil, err
		}

		if status == nil {
			// there is no status, lets deploy
			deployment := Deployment{
				SHA: *gh_deployment.SHA,
				ID:  *gh_deployment.ID,
			}
			return &deployment, nil
		}

		// TODO: Are there stuck pending statuses?
		// TODO: Allow for commands to run only once
	}

	// No deployment for this environment
	return nil, nil
}

type StatusDescription struct {
	Server string
}

// UpdateStatus
// Creates a deployment status in the backend with the given status to a given deployment
func (b *GithubBackend) UpdateStatus(deployment *Deployment, status string) error {
	client := b.getClient()
	user, repo := b.parseGithubInfo()

	desc, _ := json.Marshal(StatusDescription{Server: b.Config.ServerId})

	req := github.DeploymentStatusRequest{
		State:       &status,
		Description: github.String(string(desc)),
	}
	_, _, err := client.Repositories.CreateDeploymentStatus(user, repo, deployment.ID, &req)

	return err
}

func (b *GithubBackend) CreateDeployment(ref string, env string) error {
	client := b.getClient()
	user, repo := b.parseGithubInfo()

	status_req := github.DeploymentRequest{
		Ref:         &ref,
		Task:        github.String("deploy"),
		AutoMerge:   github.Bool(false),
		Environment: &env,
	}

	_, _, err := client.Repositories.CreateDeployment(user, repo, &status_req)

	if err != nil {
		return err
	}

	return nil
}

// Get the server status for a given deployment
func (b *GithubBackend) getLatestStatus(deployment *github.Deployment) (*github.DeploymentStatus, error) {
	client := b.getClient()
	user, repo := b.parseGithubInfo()

	// find all statuses
	statuses, _, err := client.Repositories.ListDeploymentStatuses(user, repo, *deployment.ID, nil)
	if err != nil {
		return nil, err
	}

	// Are there statuses?
	if len(statuses) > 0 {
		// find the latest status for this server
		for _, status := range statuses {
			var desc StatusDescription
			json.Unmarshal([]byte(*status.Description), &desc)
			if desc.Server == b.Config.ServerId {
				return &status, nil
			}
		}
	}

	// No relevant status found
	return nil, nil
}

// Extract the username and repo from the github url
func (b *GithubBackend) parseGithubInfo() (string, string) {
	r := strings.NewReplacer("https://github.com/", "", ".git", "")
	repo_info := strings.Split(r.Replace(b.Config.GitUrl), "/")

	return repo_info[0], repo_info[1]
}
