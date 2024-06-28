package gosume

import (
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"gopkg.in/ini.v1"
)

type gosumeConfig struct {
	configFile     *ini.File
	credentialFile *ini.File
}

// type profileConfig struct {
// 	name          string
// 	sourceProfile string
// 	roleArn       string
// 	mfaSerial     string
// }

func newGosumeConfig() (*gosumeConfig, error) {
	conf, err := ini.Load(config.DefaultSharedConfigFilename())
	if err != nil {
		return nil, err
	}
	creds, err := ini.Load(config.DefaultSharedCredentialsFilename())
	if err != nil {
		return nil, err
	}

	return &gosumeConfig{
		configFile:     conf,
		credentialFile: creds,
	}, nil
}

func (gc *gosumeConfig) hasProfile(profile string) bool {
	return gc.configFile.HasSection("profile " + profile)
}

func (gc *gosumeConfig) mfaSerial(profile string) string {
	if !gc.hasProfile(profile) {
		return ""
	}
	section := gc.credentialFile.Section(profile)

	mfa_serial := section.Key("mfa_serial")
	if mfa_serial == nil {
		return ""
	}

	return mfa_serial.String()
}

func (gc *gosumeConfig) sourceProfile(profile string) string {
	if !gc.hasProfile(profile) {
		return ""
	}

	section := gc.configFile.Section("profile " + profile)
	source_profile := section.Key("source_profile")
	if source_profile == nil {
		return ""
	}

	return source_profile.String()
}

func (gc *gosumeConfig) credentials(profile string) *types.Credentials {
	if !gc.hasProfile(profile) {
		return nil
	}
	section := gc.credentialFile.Section(profile)

	accessKeyId := section.Key("aws_access_key_id").String()
	secretAccessKey := section.Key("aws_secret_access_key").String()
	if accessKeyId == "" || secretAccessKey == "" {
		return nil
	}

	return &types.Credentials{
		AccessKeyId:     &accessKeyId,
		SecretAccessKey: &secretAccessKey,
	}
}

func (gc *gosumeConfig) roleArn(profile string) string {
	if !gc.hasProfile(profile) {
		return ""
	}
	section := gc.configFile.Section("profile " + profile)

	roleArn := section.Key("role_arn")
	if roleArn == nil {
		return ""
	}

	return roleArn.String()
}
