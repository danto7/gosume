package gosume

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/charmbracelet/log"
)

type profileChain []string

func (g *Gosume) getRoleChain(targetProfile string) *profileChain {
	profileChain := make(profileChain, 0)
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

	conf, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithSharedConfigProfile(profileName))
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
	if mfaDevice := g.config.mfaSerial(profileName); mfaDevice != "" {
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

func readTokenCode() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter TOTP: ")
	text, err := reader.ReadString('\n')

	return strings.TrimSpace(text), err
}

func createConfigFromStaticCredentials(ctx context.Context, session *types.Credentials) (*aws.Config, error) {
	conf, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     *session.AccessKeyId,
				SecretAccessKey: *session.SecretAccessKey,
				SessionToken:    *session.SessionToken,
			},
		}))

	return &conf, err
}
