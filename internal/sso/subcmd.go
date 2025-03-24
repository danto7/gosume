package sso

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/charmbracelet/log"
	"github.com/danto7/gosume/internal/cache"
	"github.com/danto7/gosume/internal/config"
	"github.com/go-errors/errors"
	"github.com/urfave/cli/v3"
	"gopkg.in/ini.v1"
)

var ErrorDeviceAuthorizationTimeout = errors.Errorf("device authorization timeout")

func Subcmd(ctx context.Context, cmd *cli.Command, conf *config.Config) error {
	cache, err := cache.NewCredentialsCache()
	if err != nil {
		log.Error("could not create credentials cache", "error", err)
		return err
	}
	sessionName := cmd.Args().First()
	log.Debug("parameters", "sessionName", sessionName)
	if sessionName == "" {
		return errors.Errorf("no sso session")
	}

	cachedSession := cache.CachedSsoSession(sessionName)
	// TODO: Check if session is expired and refresh
	if cachedSession != nil && cachedSession.Expired() {
		log.Debug("cached sso session expired, attempting refresh")

		return errors.Errorf("session expired")
	}
	if cachedSession == nil {
		log.Debug("no cached sso session found, starting device authorization", "session", sessionName)

		section := conf.SsoSession(sessionName)
		if section == nil {
			return errors.Errorf("could not find sso session %s", sessionName)
		}

		cachedSession, err = deviceAuth(ctx, sessionName, section, cache)
		if err != nil {
			return errors.Errorf("device auth failed: %w", err)
		}
	}

	log.Debug("using sso session", "session", sessionName, "expires_at", cachedSession.ExpiresAt)
	accounts, err := cachedSession.Accounts(ctx)
	if err != nil {
		return errors.Errorf("could not create sso client: %w", err)
	}

	for accounts.HasMorePages() {
		page, err := accounts.NextPage(ctx)
		if err != nil {
			return errors.Errorf("could not list accounts: %w", err)
		}
		for _, account := range page.AccountList {
			fmt.Println(*account.AccountName)
		}
	}

	return nil
}

func deviceAuth(ctx context.Context, sessionName string, sessionConfig *ini.Section, creds *cache.CredentialsCache) (*cache.CachedSsoSession, error) {
	region, err := sessionConfig.GetKey("sso_region")
	if err != nil {
		return nil, errors.Errorf("could not find sso region for session %s: %w", sessionName, err)
	}
	startUrl, err := sessionConfig.GetKey("sso_start_url")
	if err != nil {
		return nil, errors.Errorf("could not find sso start url for session %s: %w", sessionName, err)
	}

	cfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(region.Value()))
	client := ssooidc.NewFromConfig(cfg)

	input := ssooidc.StartDeviceAuthorizationInput{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		StartUrl:     aws.String(startUrl.Value()),
	}
	deviceAuth, err := client.StartDeviceAuthorization(ctx, &input)
	if err != nil {
		return nil, err
	}
	log.Debug("device auth started", "device code", *deviceAuth.DeviceCode, "verification uri", *deviceAuth.VerificationUriComplete)

	deviceAuthCtx, deviceAuthCancel := context.WithTimeout(ctx, time.Duration(deviceAuth.ExpiresIn)*time.Second)
	createTokenInput := ssooidc.CreateTokenInput{
		GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"),
		ClientId:     clientId,
		ClientSecret: clientSecret,
		DeviceCode:   deviceAuth.DeviceCode,
	}
	var createTokenOut *ssooidc.CreateTokenOutput
	for {
		out, err := client.CreateToken(deviceAuthCtx, &createTokenInput)
		if err == nil {
			log.Debug("poll token succeeded")
			createTokenOut = out
			break
		}

		log.Debug("poll token failed", "error", err)
		// TODO: Check for authorization pending and handle other kind of errors differently
		select {
		case <-time.After(time.Duration(deviceAuth.Interval) * time.Second):
			continue
		case <-deviceAuthCtx.Done():
			deviceAuthCancel()
			return nil, ErrorDeviceAuthorizationTimeout
		}
	}
	deviceAuthCancel()
	log.Debug("device auth successfull", "response", fmt.Sprintf("%+v", createTokenOut))

	expiresAt := time.Now().Add(time.Duration(createTokenOut.ExpiresIn) * time.Second)
	session := cache.CachedSsoSession{
		AccessToken:  createTokenOut.AccessToken,
		RefreshToken: createTokenOut.RefreshToken,
		ExpiresAt:    &expiresAt,
		SsoRegion:    aws.String(region.Value()),
	}
	creds.SetSsoSession(sessionName, &session)
	if err := creds.Write(); err != nil {
		return nil, err
	}

	return &session, nil
}
