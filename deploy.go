package main

import (
	"bytes"
	// "errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"

	"github.com/google/go-github/github"
)

type ByDate []os.FileInfo

func (a ByDate) Len() int           { return len(a) }
func (a ByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool { return a[j].ModTime().After(a[i].ModTime()) }

func Deploy(config *Config) error {
	deployment, err := findNewDeployment(config)
	if err != nil {
		return err
	}

	if deployment == nil {
		log.Println("No new deployment")
		return nil
	}

	log.Println("No status found...")

	// Lets ping the api before deploying
	log.Println("Setting to pending")
	createDeployStatus(config, deployment, "pending")

	// Actually deploy
	checkoutPath, err := doCheckout(config, deployment)
	if err != nil {
		log.Println("There was an error checking out the code")
		createDeployStatus(config, deployment, "error")
		return err
	}

	// Symlink Shared Files
	log.Println("Symlink Shared Files")
	if err = doSharedSymlink(config, checkoutPath); err != nil {
		createDeployStatus(config, deployment, "error")
		return err
	}

	// Run Before Symlink tasks
	log.Println("Before Symlink")
	if err = doSymlinkStep(config, checkoutPath, true); err != nil {
		createDeployStatus(config, deployment, "error")
		return err
	}

	// Symlink /current to last release
	log.Println("Symlink")
	if err = doSymlink(config, checkoutPath); err != nil {
		createDeployStatus(config, deployment, "error")
		return err
	}

	// Run After Symlink tasks
	log.Println("After Symlink")
	if err = doSymlinkStep(config, checkoutPath, false); err != nil {
		createDeployStatus(config, deployment, "error")
		return err
	}

	if err := doCleanup(config); err != nil {
		createDeployStatus(config, deployment, "error")
		return err
	}

	// Lets mark the deploy as a success
	createDeployStatus(config, deployment, "success")
	log.Println("Set to success")

	return nil
}

func doCheckout(config *Config, deployment *github.Deployment) (*string, error) {
	log.Println("Checking out code")
	checkoutPath := config.AppPath + "/releases/" + *deployment.SHA
	if _, err := os.Stat(checkoutPath); err == nil {
		return &checkoutPath, nil
	}

	// clone the repo first
	args := []string{"clone", config.GitUrl, checkoutPath}
	cmd := exec.Command("git", args...)

	// capture stderr to see if there was an issue checking out the app
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		log.Println("There was an error cloning:")
		log.Println(string(stderr.Bytes()))
		return nil, err
	}

	// get the revision we want
	cmd = exec.Command("git", "reset", "--hard", *deployment.SHA)
	cmd.Dir = checkoutPath
	cmd.Stderr = &stderr

	if err = cmd.Run(); err != nil {
		log.Println("There was an error changing rev:")
		log.Println(string(stderr.Bytes()))
		return nil, err
	}

	log.Println("Checked out")
	return &checkoutPath, nil
}

func doSharedSymlink(config *Config, checkoutPath *string) error {
	for shared, target := range config.SharedFiles {
		log.Println("Linking", shared, "to", target)
		sharedPath := config.AppPath + "/shared/" + shared
		targetPath := *checkoutPath + "/" + target

		// Remove existing file, if any
		os.RemoveAll(targetPath)

		// Do the Symlink
		err := os.Symlink(sharedPath, targetPath)
		if err != nil {
			log.Println("Error symlinking")
			log.Println(err)
			return err
		}
	}
	return nil
}

func doSymlinkStep(config *Config, checkoutPath *string, before bool) error {
	var commands []string
	if before {
		commands = config.BeforeSymlink
	} else {
		commands = config.AfterSymlink
	}

	for _, c := range commands {
		log.Println("Running: ", c)
		cmd := exec.Command("sh", "-c", c)
		cmd.Dir = *checkoutPath

		// capture stderr to see if there was an issue checking out the app
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			log.Println("Error Running: ", c)
			log.Println(string(stderr.Bytes()))
			log.Println(err)
			return err
		} else {
			log.Println("Successfully run: ", c)
		}
	}
	return nil
}

func doSymlink(config *Config, checkoutPath *string) error {
	currentPath := config.AppPath + "/current"
	err := os.RemoveAll(currentPath)
	if err != nil {
		return err
	}
	err = os.Symlink(*checkoutPath, currentPath)
	if err != nil {
		return err
	}
	return nil
}

func doCleanup(config *Config) error {
	basePath := config.AppPath + "/releases"
	files, _ := ioutil.ReadDir(basePath)

	if len(files) > config.KeepRevisions {
		sort.Sort(ByDate(files))
		removeCount := len(files) - config.KeepRevisions
		for _, d := range files[:removeCount] {
			os.RemoveAll(basePath + "/" + d.Name())
		}
	}
	return nil
}
