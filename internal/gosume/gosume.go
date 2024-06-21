package gosume

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
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

type ProfileChain []string

func (g *Gosume) getRoleChain(targetProfile string) *ProfileChain {
	profileChain := make(ProfileChain, 0)
	currentProfile := targetProfile
	for g.config.hasProfile(currentProfile) {
		profileChain = append(profileChain, currentProfile)
		sourceProfile := g.config.sourceProfile(currentProfile)
		currentProfile = sourceProfile
	}

	slices.Reverse(profileChain)
	log.Debugf("resolved profile chain: %s", profileChain)
	return &profileChain
}

func (g *Gosume) getSourceCredentials(ctx context.Context, profileName string) (*aws.Config, error) {
	if cachedProfile := g.cache.GetProfile(profileName); cachedProfile != nil {
		log.Debug("found cached credentials", "profileName", profileName)
		credentials := &types.Credentials{
			AccessKeyId:     cachedProfile.AccessKeyId,
			SecretAccessKey: cachedProfile.SecretAccessKey,
			SessionToken:    cachedProfile.SessionToken,
		}
		return createConfigFromStaticCredentials(ctx, credentials)
	}

	if g.config.credentials(profileName) == nil {
		return nil, fmt.Errorf("could not find valid long lived credentials for profile %s", profileName)
	}

	conf, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profileName))
	if err != nil {
		return nil, fmt.Errorf("could not load config for profile '%s': \n\t%w", profileName, err)
	}
	// creds, err := conf.Credentials.Retrieve(ctx)
	// if err != nil {
	// 	log.Fatalf("could not retrieve credentials for profile %s: %v", profileName, err)
	// }
	// log.Debug("Loaded config for profile", "profileName", profileName, "accessKeyId", creds.AccessKeyID)
	stsClient := sts.NewFromConfig(conf)

	sessionTokenInput := sts.GetSessionTokenInput{}
	if mfaDevice := g.config.mfaDevice(profileName); mfaDevice != "" {
		token, err := readTokenCode()
		if err != nil {
			return nil, fmt.Errorf("could not to read totp: \n\t%w", err)
		}

		sessionTokenInput.SerialNumber = &mfaDevice
		sessionTokenInput.TokenCode = &token
	}

	session, err := stsClient.GetSessionToken(ctx, &sessionTokenInput)
	if err != nil {
		return nil, fmt.Errorf("could not create temporary credentials for %s: \n\t%w", profileName, err)
	}

	g.cache.setCredentials(profileName, "eu-central-1", session.Credentials)

	return createConfigFromStaticCredentials(ctx, session.Credentials)
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

func readTokenCode() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter TOTP: ")
	text, err := reader.ReadString('\n')

	return strings.TrimSpace(text), err
}

func createConfigFromStaticCredentials(ctx context.Context, session *types.Credentials) (*aws.Config, error) {
	conf, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     *session.AccessKeyId,
				SecretAccessKey: *session.SecretAccessKey,
				SessionToken:    *session.SessionToken,
			},
		}))

	return &conf, err
}

type Profile struct {
	Name       string
	Credential *Credential
	Config     *Config
}

type Credential struct {
	MfaSerial string
}

type Config struct {
	RoleArn string
}
