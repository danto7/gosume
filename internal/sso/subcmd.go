package sso

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/danto7/gosume/internal/cache"
	"github.com/danto7/gosume/internal/config"
	"github.com/go-errors/errors"
	"github.com/urfave/cli/v3"
)

func Subcmd(ctx context.Context, cmd *cli.Command, conf *config.Config) error {
	cache, err := cache.Load()
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
	if cachedSession != nil && cachedSession.Expired() {
		err := refreshToken(ctx, cachedSession)
		if err != nil {
			log.Debug("could not refresh sso session", "error", err)
			log.Debug("clearing cache", "sessionName", sessionName)
			cache.SetSsoSession(sessionName, nil)
			cachedSession = nil
		}
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

	log.Debug("using sso session", "session", sessionName, "expiresAt", cachedSession.ExpiresAt)
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
