package sso

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/charmbracelet/log"
	"github.com/danto7/gosume/internal/cache"
	"github.com/go-errors/errors"
)

func refreshToken(ctx context.Context, session *cache.CachedSsoSession) error {
	log.Debug("cached sso session expired, attempting refresh")
	cfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(*session.SsoRegion))
	if err != nil {
		return errors.Errorf("could not load aws config: %w", err)
	}
	client := ssooidc.NewFromConfig(cfg)
	input := &ssooidc.CreateTokenInput{
		ClientId:     clientId,
		ClientSecret: clientSecret,
		GrantType:    aws.String("refresh_token"),
		RefreshToken: session.RefreshToken,
		Scope:        []string{"sso:account:access"},
	}
	out, err := client.CreateToken(ctx, input)
	if err != nil {
		log.Debug("session refresh failed", "error", err)
		return errors.Errorf("could not refresh sso session: %w", err)
	}
	expiresAt := time.Now().Add(time.Duration(out.ExpiresIn) * time.Second)
	session.AccessToken = out.AccessToken
	session.ExpiresAt = &expiresAt
	session.RefreshToken = out.RefreshToken

	return nil
}
