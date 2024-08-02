package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/danto7/gosume/internal/subcmds"
	"github.com/go-errors/errors"

	"github.com/urfave/cli/v2"
)

func main() {
	if log.GetLevel() == log.DebugLevel {
		log.SetReportCaller(true)
	} else {
		log.SetReportTimestamp(false)
	}
	app := &cli.App{
		Name:  "gosume",
		Usage: "make awsume even greater",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:   "shell-wrapper",
				Usage:  "is used by the shell wrapper to indicate that we are called by it",
				Hidden: true,
			},
			&cli.BoolFlag{
				Name:   "print-secrets",
				Usage:  "is used by the shell wrapper to load new environment variables",
				Hidden: true,
			},
			// TODO(danto7) I think a way to integrate with the aws cli credential_process
			// would be nice to have.
		},
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
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println("wurst")
		if err, ok := err.(*errors.Error); ok {
			fmt.Println("wurst2")
			fmt.Println(err.ErrorStack())
		} else {
			fmt.Println("wurst3")
			log.Fatal(err)
		}
	}
}
