package main

import (
	"code.google.com/p/goauth2/oauth"
	"github.com/codegangsta/cli"
	"github.com/dlapiduz/go-github/github"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	GitUrl      string `yaml:"git_url"`
	Environment string
	AppPath     string `yaml:"app_path"`
	Revision    string
	ServerId    string `yaml:"server_id"`
}

func (c *Config) GetGithubClient() *github.Client {
	gh_key := os.Getenv("GH_KEY")

	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: gh_key},
	}

	client := github.NewClient(t.Client())

	return client
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

func LoadConfig(context *cli.Context) Config {
	c := Config{}

	if context.String("config") != "" {
		c = ParseConfig(context.String("config"))
		c.AppPath, _ = filepath.Abs(c.AppPath)
	} else {
		if context.String("app-path") != "" {
			c.AppPath = context.String("app-path")
		} else {
			c.AppPath, _ = filepath.Abs(".")
		}
	}
	return c

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
	app.Action = func(c *cli.Context) {
		println("boom! I say!")
	}

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

	app.Commands = []cli.Command{
		{
			Name:  "deploy",
			Usage: "Check for new deployments and execute them",
			Action: func(context *cli.Context) {
				Deploy(context)
			},
			Flags: globalFlags,
		},
		{
			Name:  "setup",
			Usage: "Create folder structure for deployments",
			Action: func(context *cli.Context) {
				Setup(context)
			},
			Flags: globalFlags,
		},
		{
			Name:  "test",
			Usage: "Lets test some stuff",
			Action: func(context *cli.Context) {
				Test(context)
			},
			Flags: globalFlags,
		},
	}

	app.Run(os.Args)
}
