package subcmds

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/danto7/gosume"
	"github.com/go-errors/errors"
	"github.com/urfave/cli/v2"
)

func Execute(cliCtx *cli.Context) error {
	ctx := cliCtx.Context
	profileName := cliCtx.Args().First()
	command := cliCtx.Args().Tail()
	log.Debug("exec", "profileName", profileName, "command", command)
	if len(command) <= 0 {
		return errors.Errorf("a command is required")
	}

	gs, err := gosume.New()
	if err != nil {
		return errors.Errorf("could not initialize gosume: \n%w", err)
	}
	cfg, err := gs.AssumeProfile(ctx, profileName)
	if err != nil {
		return errors.Errorf("could not assume profile: \n%w", err)
	}
	credentials, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return errors.Errorf("could not get credentials: \n%w", err)
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
