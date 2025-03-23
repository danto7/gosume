package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/charmbracelet/log"
	"github.com/go-errors/errors"
	"gopkg.in/ini.v1"
)

type Config struct {
	awsConfigFileName      string
	awsCredentialsFileName string
	awsConfig              *ini.File
}

func New() (*Config, error) {
	envConfig, err := config.NewEnvConfig()
	if err != nil {
		return nil, errors.Errorf("could not load aws env config: %w", err)
	}

	sharedConfigFile := envConfig.SharedConfigFile
	if sharedConfigFile == "" {
		config.DefaultSharedConfigFilename()
	}
	sharedCredentialsFile := envConfig.SharedCredentialsFile
	if sharedCredentialsFile == "" {
		sharedCredentialsFile = config.DefaultSharedCredentialsFilename()
	}
	log.Debug("loaded aws sdk env config", "config file", sharedConfigFile, "credentials file", sharedCredentialsFile)

	configIni, err := ini.Load(sharedConfigFile)
	if err != nil {
		return nil, errors.Errorf("could not load ini: %w", err)
	}

	return &Config{
		awsConfigFileName:      sharedConfigFile,
		awsCredentialsFileName: sharedCredentialsFile,
		awsConfig:              configIni,
	}, nil
}

func (gc *Config) AwsSharedConfig(profile string) (config.SharedConfig, error) {
	return config.LoadSharedConfigProfile(context.TODO(), profile,
		func(lsco *config.LoadSharedConfigOptions) {
			lsco.ConfigFiles = []string{gc.awsConfigFileName}
			lsco.CredentialsFiles = []string{gc.awsCredentialsFileName}
		})
}

func (gc *Config) HasProfile(profile string) bool {
	return gc.awsConfig.HasSection("profile " + profile)
}

func (gc *Config) SourceProfile(profile string) string {
	sharedConfig, err := gc.AwsSharedConfig(profile)
	log.Debug(err)
	if err != nil {
		return ""
	}

	return sharedConfig.SourceProfileName
}

func (gc *Config) Credentials(profile string) *types.Credentials {
	sharedConfig, err := gc.AwsSharedConfig(profile)
	if err != nil || !sharedConfig.Credentials.HasKeys() {
		return nil
	}

	return &types.Credentials{
		AccessKeyId:     &sharedConfig.Credentials.AccessKeyID,
		SecretAccessKey: &sharedConfig.Credentials.SecretAccessKey,
	}
}

func (gc *Config) RoleArn(profile string) string {
	sharedConfig, err := gc.AwsSharedConfig(profile)
	if err != nil {
		return ""
	}

	return sharedConfig.RoleARN
}

func (gc *Config) SsoSession(sessionName string) *ini.Section {
	log.Debug("config sections", "sections", fmt.Sprintf("%+v", gc.awsConfig.SectionStrings()))
	for _, section := range gc.awsConfig.Sections() {
		if strings.HasPrefix(strings.ToLower(section.Name()), "sso-session") &&
			strings.HasSuffix(section.Name(), sessionName) {
			return section
		}
	}

	log.Debug("could not find sso session '%s'", sessionName)
	return nil
}
