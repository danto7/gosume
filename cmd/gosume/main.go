package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/charmbracelet/log"
	"github.com/danto7/gosume/internal/gosume"
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
		Name:   "gosume",
		Usage:  "Make Awsume even greater",
		Action: assume,
		Commands: []*cli.Command{
			{
				Name:        "exec",
				Aliases:     []string{"e"},
				Description: "execute a command with the give profile",
				Action:      execute,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func execute(cliCtx *cli.Context) error {
	ctx := context.TODO()
	profileName := cliCtx.Args().First()
	command := cliCtx.Args().Tail()
	log.Debug("exec", "profileName", profileName, "command", command)
	if len(command) <= 0 {
		return fmt.Errorf("a command is required")
	}

	gs, err := gosume.New()
	if err != nil {
		return fmt.Errorf("could not initialize gosume: \n%w", err)
	}
	cfg, err := gs.AssumeProfile(ctx, profileName)
	if err != nil {
		return fmt.Errorf("could not assume profile: \n%w", err)
	}
	credentials, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return fmt.Errorf("could not get credentials: \n%w", err)
	}
	env := []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", credentials.AccessKeyID),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", credentials.SecretAccessKey),
		fmt.Sprintf("AWS_SESSION_TOKEN=%s", credentials.SessionToken),
		fmt.Sprintf("AWS_REGION=%s", "eu-central-1"),
	}

	inheritedEnv := os.Environ()
	inheritedEnv = slices.DeleteFunc(inheritedEnv, func(elem string) bool {
		return strings.HasPrefix(elem, "AWS_")
	})

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Env = append(inheritedEnv, env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	log.Debug("run command", "cmd", cmd)
	return cmd.Run()
}

func assume(cliCtx *cli.Context) error {
	profileName := cliCtx.Args().First()
	if profileName == "" {
		return cli.Exit("Please provide a profile name", 1)
	}

	gs, err := gosume.New()
	if err != nil {
		return fmt.Errorf("could not create gosume: %w", err)
	}

	ctx := context.TODO() // TODO: handle Ctrl-C properly
	config, err := gs.AssumeProfile(ctx, profileName)
	if err != nil {
		return fmt.Errorf("could not assume profile: %w", err)
	}
	stsClient := sts.NewFromConfig(*config)

	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("could not get caller identity: %w", err)
	}
	log.Debug("Identity", "userid", *identity.UserId, "account", *identity.Account, "arn", *identity.Arn)

	return nil
}
