package gosume

import (
	"testing"

	"gopkg.in/ini.v1"
)

func setupGosumeConfig(config_ini_raw, credential_ini_raw string) (*gosumeConfig, error) {
	config_ini, err := ini.Load([]byte(config_ini_raw))
	if err != nil {
		return nil, err
	}
	credential_ini, err := ini.Load([]byte(credential_ini_raw))
	if err != nil {
		return nil, err
	}

	return &gosumeConfig{
		configFile:     config_ini,
		credentialFile: credential_ini,
	}, nil

}

func TestHasProfileExists(t *testing.T) {
	config, err := setupGosumeConfig(`
	[profile exists]
	`, ``)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	got := config.hasProfile("exists")
	expected := true
	if got != expected {
		t.Errorf("profile 'exists' should exist")
	}
}

func TestHasProfileDoesNotExist(t *testing.T) {
	config, err := setupGosumeConfig(``, ``)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	got := config.hasProfile("nothing")
	expected := false
	if got != expected {
		t.Errorf("profile 'nothing' should not exist")
	}
}

func TestMfaDeviceProfileExistsMfaSet(t *testing.T) {
	config, err := setupGosumeConfig(`
	[profile mfa]
	`, `
	[mfa]
	mfa_serial = you_can_call_me_mfa
	`)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	if config.mfaSerial("mfa") != "you_can_call_me_mfa" {
		t.Errorf("there should have been a mfa")
	}
}

func TestMfaDeviceProfileExistsMfaNotSet(t *testing.T) {
	config, err := setupGosumeConfig(`
	[profile mfa]
	`, `
	[mfa]
	`)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}
	if config.mfaSerial("exists") != "" {
		t.Errorf("this is not the mfa_serial you are looking for")
	}
}
func TestMfaDeviceProfileDoesNotExist(t *testing.T) {
	config, err := setupGosumeConfig(`
	`, `
	[mfa]
	mfa_serial = you_can_call_me_mfa
	`)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}
	if config.mfaSerial("nothing") != "" {
		t.Errorf("this is not the mfa_serial you are looking for")
	}
}

func TestSourceProfileSet(t *testing.T) {
	config, err := setupGosumeConfig(`
	[profile A]
	[profile B]
	source_profile = A
	`, `
	`)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	if config.sourceProfile("B") != "A" {
		t.Errorf("should have been A")
	}
}

func TestSourceProfileNotSet(t *testing.T) {
	config, err := setupGosumeConfig(`
	[profile A]
	[profile B]
	`, `
	`)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	if config.sourceProfile("B") != "" {
		t.Errorf("should not been set")
	}
}

func TestCredentialsHappyCase(t *testing.T) {
	config, err := setupGosumeConfig(`
	[profile A]
	`, `
	[A]
	aws_access_key_id = wurst
	aws_secret_access_key = käse
	`)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	creds := config.credentials("A")
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

func TestCredentialsHasNoProfile(t *testing.T) {
	config, err := setupGosumeConfig(`
	`, `
	[A]
	aws_access_key_id = wurst
	aws_secret_access_key = käse
	`)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	creds := config.credentials("A")
	if creds != nil {
		t.Errorf("should be nil")
	}
}

func TestCredentialsHasNoAccessKeyId(t *testing.T) {
	config, err := setupGosumeConfig(`
	[profile A]
	`, `
	[A]
	aws_secret_access_key = käse
	`)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	creds := config.credentials("A")
	if creds != nil {
		t.Errorf("should be nil")
	}
}

func TestCredentialsHasNoSecretAccessKey(t *testing.T) {
	config, err := setupGosumeConfig(`
	[profile A]
	`, `
	[A]
	aws_access_key_id = wurst
	`)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	creds := config.credentials("A")
	if creds != nil {
		t.Errorf("should be nil")
	}
}

func TestCredentialsHasNoCredentials(t *testing.T) {
	config, err := setupGosumeConfig(`
	[profile A]
	`, `
	`)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	creds := config.credentials("A")
	if creds != nil {
		t.Errorf("should be nil")
	}
}
