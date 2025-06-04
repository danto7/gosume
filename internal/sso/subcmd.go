package sso

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sso/types"
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
	defer func() {
		if err := cache.Write(); err != nil {
			log.Error("could not write credentials cache", "error", err)
		}
	}()

	// gosume sso testlz ipsec Admin
	sessionName := cmd.Args().Get(0)
	log.Debug("parameters", "sessionName", sessionName)
	if sessionName == "" {
		return errors.Errorf("no sso session has been selected")
	}

	cachedSession := cache.CachedSsoSession(sessionName)
	if cachedSession != nil && cachedSession.Expired() {
		log.Debug("cached sso session expired, attempting refresh", "currentTime", time.Now(), "expiresAt", cachedSession.ExpiresAt)
		err := refreshToken(ctx, cachedSession)
		if err != nil {
			log.Debug("could not refresh sso session", "error", err)
			// TODO: we might need to handle the kind of error a bit better. For Example we
			// might not want to clear the cache on a network error
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
	accountList, err := cachedSession.Accounts(ctx)
	if err != nil {
		return errors.Errorf("could not create sso client: %w", err)
	}
	log.Debug("got account list", "accountList", accountList)

	accountName := cmd.Args().Get(1)
	if accountName == "" {
		for _, account := range *accountList {
			fmt.Println(*account.AccountName)
		}
		return errors.Errorf("no account has been selected")
	}

	var accountDetails *types.AccountInfo
	for _, account := range *accountList {
		if accountName == *account.AccountName {
			accountDetails = &account
		}
	}
	if accountDetails == nil {
		return errors.Errorf("account %s not found in accountList", accountName)
	}

	roleList, err := cachedSession.Roles(ctx, accountName)
	if err != nil {
		return errors.Errorf("could not list roles: %w", err)
	}

	roleName := cmd.Args().Get(2)
	if roleName == "" {
		for _, role := range *roleList {
			fmt.Println(*role.RoleName)
		}
		return errors.Errorf("no role has been selected")
	}

	// TODO: get role credentials

	return nil
}
