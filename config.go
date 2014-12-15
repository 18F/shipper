package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"gopkg.in/yaml.v2"
)

type Config struct {
	GitUrl        string `yaml:"git_url"`
	Environment   string
	AppPath       string   `yaml:"app_path"`
	ServerId      string   `yaml:"server_id"`
	BeforeSymlink []string `yaml:"before_symlink"`
	AfterSymlink  []string `yaml:"after_symlink"`
	Interval      int
	SharedFiles   map[string]string `yaml:"shared_files"`
	KeepRevisions int               `yaml:"keep_revisions"`

	BackendName string `yaml:"backend_name"`
	Backend     Backend
}

type Backend interface {
	FindNewDeployment() (*Deployment, error)
	UpdateStatus(deployment *Deployment, status string) error
	CreateDeployment(ref string, env string) error
}

type Deployment struct {
	ID  int
	SHA string
}

// Checks if a config file is present and loads it
func LoadConfig(context *cli.Context) (*Config, error) {
	// Set defaults for config
	c := Config{
		KeepRevisions: 3,
		BackendName:   "github",
	}

	if context.String("config") == "" {
		return nil, errors.New("No config path provided")
	}
	if os.Getenv("GH_KEY") == "" {
		return nil, errors.New("GH_KEY is a required env variable")
	}

	data, err := ioutil.ReadFile(context.String("config"))
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	// Make sure app path is absolute
	c.AppPath, _ = filepath.Abs(c.AppPath)

	if c.BackendName == "github" {
		backend := GithubBackend{Config: &c}
		c.Backend = &backend
	}

	return &c, nil
}
