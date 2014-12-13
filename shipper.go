package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Jack the shipper"
	app.Usage = "Continuous deployment made easy and secure"
	app.Version = "0.2.0"

	globalFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "config path",
		},
	}
	newFlags := []cli.Flag{
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
				config, err := LoadConfig(context)
				if err != nil {
					return
				}
				Setup(config)
			},
			Flags: globalFlags,
		},
		{
			Name:  "new",
			Usage: "Create new deployment",
			Action: func(context *cli.Context) {
				config, err := LoadConfig(context)
				if err != nil {
					return
				}
				Create(context, config)
			},
			Flags: append(globalFlags, newFlags...),
		},
		{
			Name:  "run",
			Usage: "Continuously check for deployments",
			Action: func(context *cli.Context) {
				config, err := LoadConfig(context)
				if err != nil {
					return
				}
				Run(config)
			},
			Flags: globalFlags,
		},
	}

	app.Run(os.Args)
}
