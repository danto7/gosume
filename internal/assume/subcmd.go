package assume

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/charmbracelet/log"
	"github.com/danto7/gosume/internal/gosume"
	"github.com/go-errors/errors"
	"github.com/urfave/cli/v3"
)

var ShellWrapperUpdateEnvironment = cli.Exit("", 42)

func Subcmd(ctx context.Context, cmd *cli.Command) error {
	profileName := cmd.Args().First()
	if profileName == "" {
		return cli.Exit("Please provide a profile name", 1)
	}

	gs, err := gosume.New()
	if err != nil {
		return errors.Errorf("could not create gosume: %w", err)
	}

	config, err := gs.AssumeProfile(ctx, profileName)
	if err != nil {
		return errors.Errorf("could not assume profile: %w", err)
	}

	if !cmd.Bool("print-secrets") {
		log.Infof("Assumed profile %s", profileName)
	}

	if cmd.Bool("shell-wrapper") {
		if cmd.Bool("print-secrets") {
			credentials, err := config.Credentials.Retrieve(ctx)
			if err != nil {
				return errors.Errorf("could not get credentials: \n%w", err)
			}
			fmt.Printf("AWS_ACCESS_KEY_ID=%s\n", credentials.AccessKeyID)
			fmt.Printf("AWS_SECRET_ACCESS_KEY=%s\n", credentials.SecretAccessKey)
			fmt.Printf("AWS_SESSION_TOKEN=%s\n", credentials.SessionToken)
			fmt.Printf("AWS_REGION=%s\n", "eu-central-1")

			return nil
		} else {
			return ShellWrapperUpdateEnvironment
		}
	}

	stsClient := sts.NewFromConfig(*config)

	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return errors.Errorf("could not get caller identity: %w", err)
	}
	log.Info("Identity", "userid", *identity.UserId, "account", *identity.Account, "arn", *identity.Arn)
	log.Warn("Your shell is not correctly configured for this command. See `gosume completion help` for more information")

	return nil
}
