package main_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"text/template"
)

func expect(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected %s but got: %s", message, err)
	}
}

func buildBinary(t *testing.T) string {
	t.Helper()
	outPath := t.TempDir()
	binaryPath := path.Join(outPath, "gosume")
	cmd := exec.Command("go", "build", "-o", binaryPath, "../gosume/main.go")
	wd, err := os.Getwd()
	expect(t, err, "get working directory returned no error")
	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	expect(t, cmd.Run(), "go build to succeed")

	return binaryPath
}

type ShellTestSetupTemplateParameters struct {
	BinaryPath string
	Script     string
	TestDir    string
}

const testSetupScript = `
export GOSUME_LOG_LEVEL=debug

export XDG_CONFIG_HOME='{{ .TestDir }}'
export XDG_CACHE_HOME='{{ .TestDir }}'
export XDG_DATA_HOME='{{ .TestDir }}'
export AWS_CONFIG_FILE
export AWS_SHARED_CREDENTIALS_FILE

echo --- $AWS_CONFIG_FILE ---
cat $AWS_CONFIG_FILE
echo
echo -----------------
echo --- $AWS_SHARED_CREDENTIALS_FILE ---
cat $AWS_SHARED_CREDENTIALS_FILE | sed 's/\=.*$/= ***redacted***/'
echo -----------------

{{ .Script }}
`

type integrationTestParameters struct {
	AwsAccessKeyId      string
	AwsSecretAccessKey  string
	BinaryPath          string
	AdditionalAwsConfig string
	Shell               shellType
	Script              string
}

func (i *integrationTestParameters) evalShellCompletion(t *testing.T) string {
	t.Helper()

	if i.Shell == fish {
		return fmt.Sprintf("%s completion fish | source", i.BinaryPath)
	} else if i.Shell == zsh {
		return fmt.Sprintf("eval \"$(%s completion zsh)\"", i.BinaryPath)
	}
	t.Fatalf("unsupported shell type %v", i.Shell)
	return ""
}

func (i *integrationTestParameters) createCmd(t *testing.T, scriptPath string) *exec.Cmd {
	if i.Shell == fish {
		return exec.Command("fish", "-l", "--private", scriptPath)
	} else if i.Shell == zsh {
		return exec.Command("zsh", "-eo", "pipefail", scriptPath)
	}
	t.Fatalf("unsupported shell type %v", i.Shell)
	return nil
}

type shellType int

const (
	fish shellType = iota
	zsh
)

func (shell shellType) String() string {
	if shell == fish {
		return "fish"
	} else if shell == zsh {
		return "zsh"
	}
	panic(fmt.Sprintf("unknown shell %T", shell))
}

func TestWithValidAWSCredentials(t *testing.T) {
	awsAccessKeyId := os.Getenv("GOSUME_AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("GOSUME_AWS_SECRET_ACCESS_KEY")
	if awsAccessKeyId == "" || awsSecretAccessKey == "" {
		t.Skip("there are no valid aws credentials to test assume against the aws api")
	}

	params := integrationTestParameters{
		AwsAccessKeyId:     awsAccessKeyId,
		AwsSecretAccessKey: awsSecretAccessKey,
		BinaryPath:         buildBinary(t),
	}

	for _, shell := range []shellType{fish, zsh} {
		p := params
		p.Shell = shell

		t.Run(p.Shell.String()+": assume static credentials", func(t *testing.T) {
			p.Script = fmt.Sprintf("export PATH=$PATH:%s\ngosume assume test-source",
				path.Dir(params.BinaryPath))
			output := runWithShell(t, p)
			t.Log(output)

			if !strings.Contains(output, "Assumed profile test-source") {
				t.Fail()
			}
		})
	}
}

func runWithShell(t *testing.T, params integrationTestParameters) string {
	testDir := t.TempDir()

	env := prepareAwsTestConfig(t, testDir, params)

	tmpl, err := template.New("testSetupScript").Parse(testSetupScript)
	if err != nil {
		t.Fatalf("could not parse %s template %s", tmpl.Name(), err)
	}

	var renderedScript bytes.Buffer
	err = tmpl.Execute(&renderedScript, ShellTestSetupTemplateParameters{
		BinaryPath: params.BinaryPath,
		TestDir:    testDir,
		Script:     params.evalShellCompletion(t) + "\n" + params.Script,
	})
	expect(t, err, "template execution succeeded")

	t.Logf("renderedScript: \n%s", &renderedScript)

	scriptPath := path.Join(testDir, "test.sh")
	err = os.WriteFile(scriptPath, renderedScript.Bytes(), 0700)
	expect(t, err, "writing test script to tempdir")

	var scriptOutput bytes.Buffer
	cmd := params.createCmd(t, scriptPath)
	cmd.Dir = testDir
	cmd.Stdout = &scriptOutput
	cmd.Stderr = &scriptOutput
	cmd.Env = env

	err = cmd.Run()
	if err != nil {
		t.Logf("test script: \n%s", &renderedScript)
		t.Logf("test script result: \n%s", scriptOutput.String())

		t.Fatalf("error while running script: %s", err)
	}

	return scriptOutput.String()
}

func prepareAwsTestConfig(t *testing.T, testDir string, params integrationTestParameters) []string {

	awsConfigPath := path.Join(testDir, "aws_config")
	config := strings.Join([]string{
		`[profile test-source]`,
		`region = eu-central-1`,
		``,
	}, "\n")
	err := os.WriteFile(awsConfigPath, []byte(config), 0600)
	expect(t, err, "writing aws config to test dir works")

	awsCredentialPath := path.Join(testDir, "aws_credentials")
	credentials := strings.Join([]string{
		`[test-source]`,
		fmt.Sprintf("aws_access_key_id = %s", params.AwsAccessKeyId),
		fmt.Sprintf("aws_secret_access_key = %s", params.AwsSecretAccessKey),
		``,
	}, "\n")
	err = os.WriteFile(awsCredentialPath, []byte(credentials), 0600)
	expect(t, err, "writing aws credential to test dir works")

	return []string{
		fmt.Sprintf("AWS_CONFIG_FILE=%s", awsConfigPath),
		fmt.Sprintf("AWS_SHARED_CREDENTIALS_FILE=%s", awsCredentialPath),
	}
}
