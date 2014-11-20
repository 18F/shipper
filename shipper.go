package main

import (
	"code.google.com/p/goauth2/oauth"
	"errors"
	"github.com/codegangsta/cli"
	"github.com/dlapiduz/go-github/github"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	GitUrl        string `yaml:"git_url"`
	Environment   string
	AppPath       string   `yaml:"app_path"`
	ServerId      string   `yaml:"server_id"`
	BeforeSymlink []string `yaml:"before_symlink"`
	AfterSymlink  []string `yaml:"after_symlink"`
	Interval      int
	GithubClient  *github.Client
	SharedFiles   map[string]string `yaml:"shared_files"`
	KeepRevisions int               `yaml:"keep_revisions"`
}

func (c *Config) GetGithubClient() *github.Client {
	if c.GithubClient != nil {
		return c.GithubClient
	}
	gh_key := os.Getenv("GH_KEY")

	if gh_key != "" {
		t := &oauth.Transport{
			Token: &oauth.Token{AccessToken: gh_key},
		}
		c.GithubClient = github.NewClient(t.Client())
	} else {
		c.GithubClient = github.NewClient(nil)
	}

	return c.GithubClient
}

func (c *Config) ParseGithubInfo() (string, string) {
	r := strings.NewReplacer("https://github.com/", "", ".git", "")
	repo_info := strings.Split(r.Replace(c.GitUrl), "/")

	return repo_info[0], repo_info[1]
}

func PanicOn(err error) {
	if err != nil {
		panic(err)
	}
}

func LoadConfig(context *cli.Context) (Config, error) {
	c := Config{}

	if context.String("config") != "" {
		c = ParseConfig(context.String("config"))
		c.AppPath, _ = filepath.Abs(c.AppPath)
	} else {
		return c, errors.New("No config provided")
	}

	return c, nil

}

func ParseConfig(config_path string) Config {
	c := Config{}

	data, err := ioutil.ReadFile(config_path)
	PanicOn(err)

	err = yaml.Unmarshal(data, &c)
	PanicOn(err)

	return c
}

func main() {
	app := cli.NewApp()
	app.Name = "Jack the shipper"
	app.Usage = "Continuous deployment made easy and secure"
	app.Version = "0.0.2"

	globalFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "config path",
		},
		cli.StringFlag{
			Name:  "app-path, p",
			Usage: "base path for the app",
		},
	}
	runFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "ref",
			Usage: "Deploy reference",
		},
		cli.StringFlag{
			Name:  "environment, e",
			Usage: "Deploy environment",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "setup",
			Usage: "Create folder structure for deployments",
			Action: func(context *cli.Context) {
				Setup(context)
			},
			Flags: globalFlags,
		},
		{
			Name:  "new",
			Usage: "Create new deployment",
			Action: func(context *cli.Context) {
				Create(context)
			},
			Flags: append(globalFlags, runFlags...),
		},
		{
			Name:  "run",
			Usage: "Continuously check for deployments",
			Action: func(context *cli.Context) {
				Run(context)
			},
			Flags: globalFlags,
		},
		{
			Name:  "deploy",
			Usage: "Manually check and run a deployment",
			Action: func(context *cli.Context) {
				config, _ := LoadConfig(context)
				Deploy(&config)

			},
			Flags: globalFlags,
		},
	}

	app.Run(os.Args)
}
