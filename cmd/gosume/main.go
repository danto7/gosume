package main

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/danto7/gosume/internal/subcmds"

	"github.com/urfave/cli/v2"
)

func main() {
	if os.Args[0] != "gosume" {
		log.SetLevel(log.DebugLevel)
		log.Debugf("detected development environment. Switch to debug log level")
	}
	if log.GetLevel() == log.DebugLevel {
		log.SetReportCaller(true)
	}
	// TODO: gosume "profile-name" --session-name "session-name"
	app := &cli.App{
		Name:                 "gosume",
		Usage:                "make awsume even greater",
		DefaultCommand:       "assume",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:        "completion",
				Description: "print the shell completion code for the specified shell",
				Action:      subcmds.Completion,
				ArgsUsage:   " <shell name>",
			},
			{
				Name:        "exec",
				Aliases:     []string{"e"},
				Description: "execute a command with the give profile",
				Action:      subcmds.Execute,
				ArgsUsage:   " <profile name> <command>",
			},
			{
				Name:        "assume",
				Aliases:     []string{"a"},
				Description: "assume the given profile",
				Action:      subcmds.Assume,
				ArgsUsage:   " <profile name>",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:   "print-secrets",
						Value:  false,
						Hidden: true,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
