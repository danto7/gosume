package config_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/log"

	"github.com/danto7/gosume/internal/config"
)

func trimWhitespace(content string) string {
	lines := strings.Split(content, "\n")
	result := ""
	for _, line := range lines {
		result += strings.TrimSpace(line) + "\n"
	}
	return result
}

func setupGosumeConfig(t *testing.T, config_ini_raw, credential_ini_raw string) *config.GosumeConfig {
	t.Helper()
	log.SetLevel(log.DebugLevel)

	var err error
	workDir := t.TempDir()
	configFile := filepath.Join(workDir, "aws_config")
	credentialsFile := filepath.Join(workDir, "aws_creds")
	err = os.WriteFile(configFile, []byte(trimWhitespace(config_ini_raw)), fs.ModePerm)
	if err != nil {
		t.Error(err)
	}
	err = os.WriteFile(credentialsFile, []byte(trimWhitespace(credential_ini_raw)), fs.ModePerm)
	if err != nil {
		t.Error(err)
	}
	t.Setenv("AWS_CONFIG_FILE", configFile)
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", credentialsFile)
	// TODO fix test setup
	conf, err := config.New()
	if err != nil {
		t.Error(err)
	}

	return &conf
}

func TestHasProfileExists(t *testing.T) {
	config := setupGosumeConfig(t, `
	[profile exists]
	`, ``)

	got := config.HasProfile("exists")
	expected := true
	if got != expected {
		t.Errorf("profile 'exists' should exist")
	}
}

func TestHasProfileDoesNotExist(t *testing.T) {
	config := setupGosumeConfig(t, ``, ``)

	got := config.HasProfile("nothing")
	expected := false
	if got != expected {
		t.Errorf("profile 'nothing' should not exist")
	}
}

func TestMfaDeviceProfileExistsMfaSet(t *testing.T) {
	config := setupGosumeConfig(t, `
	[profile mfa]
	`, `
	[mfa]
	mfa_serial = you_can_call_me_mfa
	`)

	if config.MfaSerial("mfa") != "you_can_call_me_mfa" {
		t.Errorf("there should have been a mfa")
	}
}

func TestMfaDeviceProfileExistsMfaNotSet(t *testing.T) {
	config := setupGosumeConfig(t, `
	[profile mfa]
	`, `
	[mfa]
	`)
	if config.MfaSerial("exists") != "" {
		t.Errorf("this is not the mfa_serial you are looking for")
	}
}
func TestMfaDeviceProfileDoesNotExist(t *testing.T) {
	config := setupGosumeConfig(t, `
	`, `
	[mfa]
	mfa_serial = you_can_call_me_mfa
	`)
	if config.MfaSerial("nothing") != "" {
		t.Errorf("this is not the mfa_serial you are looking for")
	}
}

func TestSourceProfileSet(t *testing.T) {
	config := setupGosumeConfig(t, `
	[profile A]
	[profile B]
	source_profile = A
	role_arn = arn:aws:iam::123456789012:role/testrole
	`, `
	[A]
	aws_access_key_id = jnfjksnj
	aws_secret_access_key = fnsjkn
	`)

	if source := config.SourceProfile("B"); source != "A" {
		t.Errorf("should have been A was: '%s'", source)
	}
}

func TestSourceProfileNotSet(t *testing.T) {
	config := setupGosumeConfig(t, `
	[profile A]
	[profile B]
	`, `
	`)

	if config.SourceProfile("B") != "" {
		t.Errorf("should not been set")
	}
}

func TestCredentialsHappyCase(t *testing.T) {
	config := setupGosumeConfig(t, `
	[profile A]
	`, `
	[A]
	aws_access_key_id = wurst
	aws_secret_access_key = käse
	`)

	creds := config.Credentials("A")
	if creds == nil {
		t.Errorf("should not be nil")
	} else {
		if creds.AccessKeyId != nil && *(creds.AccessKeyId) != "wurst" {
			t.Errorf("access_key_id does not match")
		}
		if creds.SecretAccessKey != nil && *(creds.SecretAccessKey) != "käse" {
			t.Errorf("secret_access_key does not match")
		}
	}
}

func TestCredentialsHasNoAccessKeyId(t *testing.T) {
	config := setupGosumeConfig(t, `
	[profile A]
	`, `
	[A]
	aws_secret_access_key = käse
	`)

	creds := config.Credentials("A")
	if creds != nil {
		t.Errorf("should be nil")
	}
}

func TestCredentialsHasNoSecretAccessKey(t *testing.T) {
	config := setupGosumeConfig(t, `
	[profile A]
	`, `
	[A]
	aws_access_key_id = wurst
	`)

	creds := config.Credentials("A")
	if creds != nil {
		t.Errorf("should be nil")
	}
}

func TestCredentialsHasNoCredentials(t *testing.T) {
	config := setupGosumeConfig(t, `
	[profile A]
	`, `
	`)

	creds := config.Credentials("A")
	if creds != nil {
		t.Errorf("should be nil")
	}
}
