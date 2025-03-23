package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/danto7/gosume/internal/assume"
	"github.com/danto7/gosume/internal/completion"
	"github.com/danto7/gosume/internal/config"
	"github.com/danto7/gosume/internal/execute"
	"github.com/danto7/gosume/internal/sso"
	"github.com/go-errors/errors"

	"github.com/urfave/cli/v3"
)

func main() {
	level, err := log.ParseLevel(os.Getenv("GOSUME_LOG_LEVEL"))
	if err != nil {
		level = log.InfoLevel
	}
	log.SetLevel(level)
	log.Debugf("log level set to %s", log.GetLevel())

	if log.GetLevel() == log.DebugLevel {
		log.SetReportCaller(true)
	} else {
		log.SetReportTimestamp(false)
	}
	cmd := &cli.Command{
		Name:  "gosume",
		Usage: "make awsume Go brrrr",
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
				Action:      completion.Subcmd,
				ArgsUsage:   " <shell name>",
			},
			{
				Name:        "exec",
				Aliases:     []string{"e"},
				Description: "execute a command with the given profile",
				Action:      execute.Subcmd,
				ArgsUsage:   " <profile name> <command>",
			},
			{
				Name:        "assume",
				Aliases:     []string{"a"},
				Description: "assume the given profile",
				Action:      assume.Subcmd,
				ArgsUsage:   " <profile name>",
			},
			{
				Name:        "sso",
				Aliases:     []string{"s"},
				Description: "assume the given profile using SSO",
				Action:      subcmdWithConfig(sso.Subcmd),
			},
		},
	}

	err = cmd.Run(context.Background(), os.Args)
	if err != nil {
		if err, ok := err.(*errors.Error); ok {
			fmt.Println()
			fmt.Println(err.ErrorStack())
		} else {
			log.Fatal(err)
		}
	}
}

func subcmdWithConfig(action func(context.Context, *cli.Command, *config.AwsConfig) error) cli.ActionFunc {
	config, err := config.NewConfig()
	if err != nil {
		return func(ctx context.Context, cmd *cli.Command) error {
			return errors.Errorf("could not load aws config: %w", err)
		}
	}
	return func(ctx context.Context, cmd *cli.Command) error {
		return action(ctx, cmd, config)
	}
}
