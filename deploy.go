package main

import (
	"bytes"
	"errors"
	"github.com/codegangsta/cli"
	"github.com/dlapiduz/go-github/github"
	"log"
	"os"
	"os/exec"
)

func checkNewDeployments(config *Config) error {
	deployment, err := findNewDeployment(config)
	PanicOn(err)

	if deployment != nil {
		log.Println("No status found...")

		// Lets ping the api before deploying
		log.Println("Setting to pending")
		err := createDeployStatus(config, deployment, "pending")
		PanicOn(err)

		// Actually deploy
		err = doCheckout(config, deployment)
		if err != nil {
			log.Println("There was an error checking out the code")
			statusErr := createDeployStatus(config, deployment, "error")
			PanicOn(statusErr)
			return err
		}
		// Lets mark the deploy as a success
		err = createDeployStatus(config, deployment, "success")
		PanicOn(err)
		log.Println("Set to success")
	}
	return nil
}

func doCheckout(config *Config, deployment *github.Deployment) error {
	log.Println("Checking out code")
	checkout_path := config.AppPath + "/releases/" + *deployment.SHA
	if _, err := os.Stat(checkout_path); err != nil {
		if os.IsExist(err) {
			return errors.New("Revision already checked out")
		}
	}

	// clone the repo first
	args := []string{"clone", config.GitUrl, checkout_path}
	cmd := exec.Command("git", args...)

	// capture stderr to see if there was an issue checking out the app
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		log.Println("There was an error cloning:")
		log.Println(string(stderr.Bytes()))
		return err
	}

	// get the revision we want
	cmd = exec.Command("git", "reset", "--hard", *deployment.SHA)
	cmd.Dir = checkout_path
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Println("There was an error changing rev:")
		log.Println(string(stderr.Bytes()))
		return err
	}

	log.Println("Checked out")
	return nil
}

func Deploy(context *cli.Context) {
	config := LoadConfig(context)
	revision := checkNewDeployments(&config)

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
