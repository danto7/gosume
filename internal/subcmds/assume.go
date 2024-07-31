package subcmds

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/charmbracelet/log"
	"github.com/danto7/gosume"
	"github.com/urfave/cli/v2"
)

func Assume(cliCtx *cli.Context) error {
	profileName := cliCtx.Args().First()
	if profileName == "" {
		return cli.Exit("Please provide a profile name", 1)
	}

	gs, err := gosume.New()
	if err != nil {
		return fmt.Errorf("could not create gosume: %w", err)
	}

	ctx := cliCtx.Context
	config, err := gs.AssumeProfile(ctx, profileName)
	if err != nil {
		return fmt.Errorf("could not assume profile: %w", err)
	}

	if cliCtx.Bool("print-secrets") {
		credentials, err := config.Credentials.Retrieve(ctx)
		if err != nil {
			return fmt.Errorf("could not get credentials: \n%w", err)
		}
		fmt.Printf("AWS_ACCESS_KEY_ID=%s\n", credentials.AccessKeyID)
		fmt.Printf("AWS_SECRET_ACCESS_KEY=%s\n", credentials.SecretAccessKey)
		fmt.Printf("AWS_SESSION_TOKEN=%s\n", credentials.SessionToken)
		fmt.Printf("AWS_REGION=%s\n", "eu-central-1")

		return nil
	}

	stsClient := sts.NewFromConfig(*config)

	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return fmt.Errorf("could not get caller identity: %w", err)
	}
	log.Debug("Identity", "userid", *identity.UserId, "account", *identity.Account, "arn", *identity.Arn)

	return nil
}
