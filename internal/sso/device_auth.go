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
	"github.com/go-errors/errors"
	"gopkg.in/ini.v1"
)

var ErrorDeviceAuthorizationTimeout = errors.Errorf("device authorization timeout")

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
	log.Info("device auth started", "device code", *deviceAuth.DeviceCode, "verification uri", *deviceAuth.VerificationUriComplete)

	deviceAuthCtx, deviceAuthCancel := context.WithTimeout(ctx, time.Duration(deviceAuth.ExpiresIn)*time.Second)
	createTokenInput := ssooidc.CreateTokenInput{
		GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"),
		ClientId:     clientId,
		ClientSecret: clientSecret,
		DeviceCode:   deviceAuth.DeviceCode,
		Scope:        []string{"sso:account:access"},
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
		AccessToken:  *createTokenOut.AccessToken,
		RefreshToken: *createTokenOut.RefreshToken,
		ExpiresAt:    expiresAt,
		SsoRegion:    region.Value(),
	}
	creds.SetSsoSession(sessionName, &session)
	if err := creds.Write(); err != nil {
		return nil, err
	}

	return &session, nil
}
