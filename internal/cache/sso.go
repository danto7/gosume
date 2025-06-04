package cache

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/go-errors/errors"
)

type CachedSsoSession struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	SsoRegion    string    `json:"sso_region"`
}

func (c *CachedSsoSession) Expired() bool {
	return time.Now().After(c.ExpiresAt)
}

type SsoAccountList []types.AccountInfo

func ssoClient(ctx context.Context, c *CachedSsoSession) (*sso.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(c.SsoRegion))
	if err != nil {
		return nil, err
	}

	return sso.NewFromConfig(cfg), nil
}

func (c *CachedSsoSession) Accounts(ctx context.Context) (*SsoAccountList, error) {
	ssoClient, err := ssoClient(ctx, c)
	if err != nil {
		return nil, errors.Errorf("could not create sso client: %w", err)
	}

	listAccountsInput := sso.ListAccountsInput{
		AccessToken: aws.String(c.AccessToken),
	}
	accounts := sso.NewListAccountsPaginator(ssoClient, &listAccountsInput)

	var accountList SsoAccountList
	for accounts.HasMorePages() {
		page, err := accounts.NextPage(ctx)
		if err != nil {
			return nil, errors.Errorf("could not list accounts page: %w", err)
		}
		accountList = append(accountList, page.AccountList...)
	}
	return &accountList, nil
}

type RoleList []types.RoleInfo

func (c *CachedSsoSession) Roles(ctx context.Context, accountId string) (*RoleList, error) {
	ssoClient, err := ssoClient(ctx, c)
	if err != nil {
		return nil, errors.Errorf("could not create sso client: %w", err)
	}
	input := &sso.ListAccountRolesInput{
		AccessToken: &c.AccessToken,
		AccountId:   &accountId,
	}
	roles := sso.NewListAccountRolesPaginator(ssoClient, input)

	var roleList RoleList
	for roles.HasMorePages() {
		page, err := roles.NextPage(ctx)
		if err != nil {
			return nil, errors.Errorf("could not list roles page: %w", err)
		}
		roleList = append(roleList, page.RoleList...)
	}
	return &roleList, nil
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
