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

const fishTestSetup = `
export GOSUME_LOG_LEVEL=debug

export XDG_CONFIG_HOME={{ .TestDir }}
export XDG_CACHE_HOME={{ .TestDir }}
export XDG_DATA_HOME={{ .TestDir }}
export AWS_CONFIG_FILE
export AWS_SHARED_CREDENTIALS_FILE

echo --- $AWS_CONFIG_FILE ---
cat $AWS_CONFIG_FILE
echo
echo -----------------
echo --- $AWS_SHARED_CREDENTIALS_FILE ---
cat $AWS_SHARED_CREDENTIALS_FILE | sed 's/\=.*$/= ***redacted***/'
echo -----------------

{{ .BinaryPath }} completion fish | source
{{ .Script }}
`

type integrationTestParameters struct {
	AwsAccessKeyId      string
	AwsSecretAccessKey  string
	BinaryPath          string
	AdditionalAwsConfig string
}

func isCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	err := cmd.Run()
	return err == nil
}

func TestWithValidAWSCredentials(t *testing.T) {
	awsAccessKeyId := os.Getenv("GOSUME_AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("GOSUME_AWS_SECRET_ACCESS_KEY")
	if awsAccessKeyId == "" || awsSecretAccessKey == "" {
		t.Skip("there are no valid aws credentials to test assume against the aws api")
	}
	if !isCommandAvailable("fish") {
		t.Fatal("fish shell is not available")
	}

	params := integrationTestParameters{
		AwsAccessKeyId:     awsAccessKeyId,
		AwsSecretAccessKey: awsSecretAccessKey,
		BinaryPath:         buildBinary(t),
	}

	t.Run("assume static credentials", func(t *testing.T) {
		command := "gosume assume test-source"
		output := runWithFish(t, &params, "", command)
		t.Log(output)

		if !strings.Contains(output, "Assumed profile test-source") {
			t.Fail()
		}
	})
}

func runWithFish(t *testing.T, params *integrationTestParameters, awsConfig string, script string) string {
	testDir := t.TempDir()
	tmpl, err := template.New("fishTestSetupTemplate").Parse(fishTestSetup)
	if err != nil {
		t.Fatalf("could not parse %s template %s", tmpl.Name(), err)
	}

	var fishScript bytes.Buffer
	err = tmpl.Execute(&fishScript, ShellTestSetupTemplateParameters{
		BinaryPath: params.BinaryPath,
		Script:     script,
		TestDir:    testDir,
	})
	expect(t, err, "template execution succeeded")

	fishScriptPath := path.Join(testDir, "test.fish")
	err = os.WriteFile(fishScriptPath, fishScript.Bytes(), 0700)
	expect(t, err, "writing test script to tempdir")

	awsConfigPath := path.Join(testDir, "aws_config")
	config := strings.Join([]string{
		`[profile test-source]`,
		`region = eu-central-1`,
		``,
		awsConfig,
	}, "\n")
	err = os.WriteFile(awsConfigPath, []byte(config), 0600)
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

	var scriptOutput bytes.Buffer
	cmd := exec.Command("fish", "-l", "--private", fishScriptPath)
	cmd.Dir = testDir
	cmd.Stdout = &scriptOutput
	cmd.Stderr = &scriptOutput
	cmd.Env = []string{
		fmt.Sprintf("AWS_CONFIG_FILE=%s", awsConfigPath),
		fmt.Sprintf("AWS_SHARED_CREDENTIALS_FILE=%s", awsCredentialPath),
	}
	err = cmd.Run()
	if err != nil {
		t.Logf("test script: \n%s", &fishScript)
		t.Logf("test script result: \n%s", scriptOutput.String())
		t.FailNow()
	}
	t.Log(os.ReadDir(testDir))

	return scriptOutput.String()
}
