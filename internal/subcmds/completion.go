package subcmds

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/urfave/cli/v2"
)

//go:embed shell_wrapper.fish
var fishTemplateString string
var fishTemplate *template.Template

func init() {
	tmpl, err := template.New("fish-wrapper").Parse(fishTemplateString)
	if err != nil {
		panic(err)
	}
	fishTemplate = tmpl
}

type shellCompletionParameters struct {
	BinaryPath string
}

func Completion(cliCtx *cli.Context) error {
	//ctx := cliCtx.Context

	shell := cliCtx.Args().First()
	params := shellCompletionParameters{
		BinaryPath: os.Args[0],
	}
	var err error
	if shell == "fish" {
		err = fishTemplate.Execute(os.Stdout, params)
	} else {
		return fmt.Errorf("The shell type '%s' is not supported yet.", shell)
	}

	return err
}
