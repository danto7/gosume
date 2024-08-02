package gosume

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/charmbracelet/log"
	"github.com/go-errors/errors"
)

func New() (*Gosume, error) {
	cache, err := NewCredentialsCache()
	if err != nil {
		return nil, err
	}

	conf, err := newGosumeConfig()
	if err != nil {
		return nil, err
	}

	return &Gosume{
		config: conf,
		cache:  cache,
	}, nil
}

type Gosume struct {
	config *gosumeConfig
	cache  *CredentialsCache
}

func (g *Gosume) AssumeProfile(ctx context.Context, targetProfileName string) (*aws.Config, error) {
	profileChain := g.getRoleChain(targetProfileName)

	if len(*profileChain) == 0 {
		return nil, errors.Errorf("could not find a profile chain for %s", targetProfileName)
	}

	// TODO: this needs heavy refactoring after tests
	var gosumeClient *Client
	for i, profileName := range *profileChain {
		log.Debugf("assuming profile %s", profileName)
		if i == 0 {
			c, err := g.getSourceCredentials(ctx, profileName)
			if err != nil {
				return nil, errors.Errorf("could not get source credentials: \n\t%w", err)
			}
			client := NewClientFromConfig(c)
			gosumeClient = &client
			continue
		}
		if gosumeClient == nil {
			panic("cfg should never be nil")
		}

		roleArn := g.config.roleArn(profileName)
		if roleArn == "" && i != 0 {
			return nil, errors.Errorf("Profile %s has no role_arn. Can not continue", profileName)
		}
		log.Debugf("assumed profile %s", profileName)

		credentials, err := gosumeClient.AssumeRole(ctx, roleArn)
		g.cache.setCredentials(profileName, "eu-central-1", credentials)
		if err != nil {
			return nil, errors.Errorf("could not create config from session credentials: \n%w", err)
		}
	}

	err := g.cache.writeCredentialsCacheToDisk()
	if err != nil {
		return nil, errors.Errorf("could not write credentials cache to disk: \n\t%w", err)
	}
	config := gosumeClient.Config()
	return &config, nil
}
