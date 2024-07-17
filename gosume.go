package gosume

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/charmbracelet/log"
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
		return nil, fmt.Errorf("could not find a profile chain for %s", targetProfileName)
	}

	// TODO: this needs heavy refactoring after tests
	var cfg *aws.Config
	var latestProfileCredentialExpiration *time.Time
	for i, profileName := range *profileChain {
		log.Debugf("assuming profile %s", profileName)
		if i == 0 {
			c, err := g.getSourceCredentials(ctx, profileName)
			if err != nil {
				return nil, fmt.Errorf("could not get source credentials: \n\t%w", err)
			}
			cfg = c
			continue
		}
		if cfg == nil {
			panic("cfg should never be nil")
		}

		stsClient := sts.NewFromConfig(*cfg)

		roleArn := g.config.roleArn(profileName)
		if roleArn == "" && i != 0 {
			return nil, fmt.Errorf("Profile %s has no role_arn. Can not continue", profileName)
		}
		session, err := stsClient.AssumeRole(ctx, &sts.AssumeRoleInput{
			RoleArn:         &roleArn,
			RoleSessionName: &profileName,
		})
		if err != nil {
			return nil, fmt.Errorf("could not assume roleArn %s: \n\t%w", roleArn, err)
		}
		log.Infof("[Assumed profile %s]", profileName)

		g.cache.setCredentials(profileName, "eu-central-1", session.Credentials)
		cfg, err = createConfigFromStaticCredentials(ctx, session.Credentials)
		if err != nil {
			return nil, fmt.Errorf("could not create config from session credentials: \n%w", err)
		}

		if session.Credentials.Expiration != nil {
			// TODO: this is broken. this does not consider the source profile
			latestProfileCredentialExpiration = session.Credentials.Expiration
		}

	}
	log.Debug("assume profile chain successfull", "latestProfileCredentialExpiration", latestProfileCredentialExpiration)

	err := g.cache.writeCredentialsCacheToDisk()
	if err != nil {
		return nil, fmt.Errorf("could not write credentials cache to disk: \n\t%w", err)
	}
	return cfg, nil
}
