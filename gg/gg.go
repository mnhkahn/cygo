package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "gg"
	app.Usage = "A Deploy tool written in Golang. It will works with Supervisor."
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		{
			Name:    "start",
			Usage:   "Rebuild & run in local path.",
			Aliases: []string{"s"},
			Action: func(c *cli.Context) {
				// Build
				if err := Build(); err == nil {
					// Run
					Start()
					// Watch()
				}
			},
		},
		{
			Name:    "build",
			Usage:   "Build.",
			Aliases: []string{"b"},
			Action: func(c *cli.Context) {
				// Build
				Build()
			},
		},
		{
			Name:    "deploy",
			Usage:   "Build & restart.",
			Aliases: []string{"d"},
			Action: func(c *cli.Context) {
				// Build()
				if NewGGConfig().IsGitPull {
					GitPull()
				}
				if err := Build(); err == nil {
					Pack()
					if err := Backup(); err != nil {
						log.Println("Delete file error", err)
					} else {
						Deploy()
					}
				}

				// if conf.NewGGConfig().IsNgrok {
				// 	Ngrok()
				// }
			},
		},
		{
			Name:    "pack",
			Usage:   "Pack & generate supervisor configuration file",
			Aliases: []string{"p"},
			Action: func(c *cli.Context) {
				if err := Build(); err == nil {
					// Supervisor
					Supervisor()

					// Pack
					if err := Pack(); err != nil {
						log.Println("Generate package error", err)
					} else {
						log.Printf("Pack success in %s.\n", NewGGConfig().AppPath)
					}
				}
			},
		},
	}

	app.Run(os.Args)
}
