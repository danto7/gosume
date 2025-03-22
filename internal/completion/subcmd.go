package completion

import (
	"context"
	_ "embed"
	"os"
	"text/template"

	"github.com/danto7/gosume/internal/assume"
	"github.com/go-errors/errors"
	"github.com/urfave/cli/v3"
)

//go:embed shell_wrapper.fish
var fishTemplateString string
var fishTemplate *template.Template

//go:embed shell_wrapper.zsh
var zshTemplateString string
var zshTemplate *template.Template

func init() {
	mustParseTemplate := func(name, templateString string) *template.Template {
		tmpl, err := template.New(name).Parse(templateString)
		if err != nil {
			panic(err)
		}
		return tmpl
	}

	fishTemplate = mustParseTemplate("fish-wrapper", fishTemplateString)
	zshTemplate = mustParseTemplate("zsh-wrapper", zshTemplateString)
}

type shellCompletionParameters struct {
	BinaryPath                string
	UpdateEnvironmentExitCode int
}

func Subcmd(ctx context.Context, cmd *cli.Command) error {
	shell := cmd.Args().First()
	params := shellCompletionParameters{
		BinaryPath:                os.Args[0],
		UpdateEnvironmentExitCode: assume.ShellWrapperUpdateEnvironment.ExitCode(),
	}
	var err error
	switch shell {
	case "fish":
		err = fishTemplate.Execute(os.Stdout, params)
	case "zsh":
		err = zshTemplate.Execute(os.Stdout, params)
	default:
		return errors.Errorf("the shell type '%s' is not supported yet", shell)
	}

	return err
}
