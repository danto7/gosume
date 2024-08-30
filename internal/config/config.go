package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/charmbracelet/log"
)

type GosumeConfig struct {
	envConfig *config.EnvConfig
}

func New() (GosumeConfig, error) {
	envConfig, err := config.NewEnvConfig()
	if err != nil {
		return GosumeConfig{}, err
	}
	log.Debug("loaded aws sdk env config", "configFile", envConfig.SharedConfigFile, "credentialsFile", envConfig.SharedCredentialsFile)
	return GosumeConfig{
		envConfig: &envConfig,
	}, nil
}

func (gc *GosumeConfig) sharedConfig(profile string) (config.SharedConfig, error) {
	return config.LoadSharedConfigProfile(context.TODO(), profile,
		func(lsco *config.LoadSharedConfigOptions) {
			lsco.ConfigFiles = []string{gc.envConfig.SharedConfigFile, config.DefaultSharedConfigFilename()}
			lsco.CredentialsFiles = []string{gc.envConfig.SharedCredentialsFile, config.DefaultSharedCredentialsFilename()}
		})
}

func (gc *GosumeConfig) HasProfile(profile string) bool {
	_, err := gc.sharedConfig(profile)
	return err == nil
}

func (gc *GosumeConfig) MfaSerial(profile string) string {
	sharedConfig, err := gc.sharedConfig(profile)
	if err != nil {
		return ""
	}

	return sharedConfig.MFASerial
}

func (gc *GosumeConfig) SourceProfile(profile string) string {
	sharedConfig, err := gc.sharedConfig(profile)
	log.Debug(err)
	if err != nil {
		return ""
	}

	return sharedConfig.SourceProfileName
}

func (gc *GosumeConfig) Credentials(profile string) *types.Credentials {
	sharedConfig, err := gc.sharedConfig(profile)
	if err != nil || !sharedConfig.Credentials.HasKeys() {
		return nil
	}

	return &types.Credentials{
		AccessKeyId:     &sharedConfig.Credentials.AccessKeyID,
		SecretAccessKey: &sharedConfig.Credentials.SecretAccessKey,
	}
}

func (gc *GosumeConfig) RoleArn(profile string) string {
	sharedConfig, err := gc.sharedConfig(profile)
	if err != nil {
		return ""
	}

	return sharedConfig.RoleARN
}
