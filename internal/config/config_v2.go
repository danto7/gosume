package config

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/charmbracelet/log"
	"github.com/go-errors/errors"
	"gopkg.in/ini.v1"
)

type AwsConfig struct {
	configFileName string
	config         *ini.File
}

func NewConfig() (*AwsConfig, error) {
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
	return &AwsConfig{
		configFileName: sharedConfigFile,
		config:         configIni,
	}, nil
}

func (ac *AwsConfig) SsoSession(session string) *ini.Section {
	log.Debug("config sections", "sections", fmt.Sprintf("%+v", ac.config.SectionStrings()))
	for _, section := range ac.config.Sections() {
		if strings.HasPrefix(strings.ToLower(section.Name()), "sso-session") &&
			strings.HasSuffix(section.Name(), session) {
			return section
		}
	}

	log.Debug("could not find sso session %s", session)
	return nil
}
