package cache

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
)

type CachedSsoSession struct {
	AccessToken  *string    `json:"access_token"`
	ExpiresAt    *time.Time `json:"expires_at"`
	RefreshToken *string    `json:"refresh_token"`
	SsoRegion    *string    `json:"sso_region"`
}

func (c *CachedSsoSession) Expired() bool {
	return time.Now().After(*c.ExpiresAt)
}

func (c *CachedSsoSession) Accounts(ctx context.Context) (*sso.ListAccountsPaginator, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(*c.SsoRegion))
	if err != nil {
		return nil, err
	}

	ssoClient := sso.NewFromConfig(cfg)
	listAccountsInput := sso.ListAccountsInput{
		AccessToken: c.AccessToken,
	}
	return sso.NewListAccountsPaginator(ssoClient, &listAccountsInput), nil
}

func (cc *CredentialsCache) SetSsoSession(sessionName string, session *CachedSsoSession) {
	cc.cache.SsoCache[sessionName] = session
}

func (cc *CredentialsCache) CachedSsoSession(sessionName string) *CachedSsoSession {
	if session, ok := cc.cache.SsoCache[sessionName]; ok {
		return session
	} else {
		return nil
	}
}
