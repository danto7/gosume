package cache

import (
	"encoding/json"
	"os"
	"time"

	"github.com/adrg/xdg"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/charmbracelet/log"
	"github.com/go-errors/errors"
)

const GOSUME_CACHE_FILENAME = "gosume-cache.json"

// TODO: check if *string is the right type for the fields or if we use string
type CachedProfile struct {
	AccessKeyId     *string    `json:"access_key_id"`
	SecretAccessKey *string    `json:"secret_access_key"`
	SessionToken    *string    `json:"session_token"`
	Expiration      *time.Time `json:"expiration"`
	Region          *string    `json:"region"`
}

type CredentialsCache struct {
	cachePath string
	cache     cache
}

type cache struct {
	ProfileCache map[string]*CachedProfile    `json:"profile_cache"`
	SsoCache     map[string]*CachedSsoSession `json:"sso_cache"`
}

func Load() (*CredentialsCache, error) {
	cachePath, err := xdg.CacheFile(GOSUME_CACHE_FILENAME)
	if err != nil {
		return nil, errors.Errorf("could not determine cache location: %w", err)
	}
	log.Debug("found cache file", "filename", cachePath)

	cc := CredentialsCache{
		cachePath: cachePath,
		cache: cache{
			ProfileCache: make(map[string]*CachedProfile),
			SsoCache:     make(map[string]*CachedSsoSession),
		},
	}
	content, err := os.ReadFile(cachePath)
	if os.IsNotExist(err) {
		return &cc, nil
	} else if err != nil {
		return nil, errors.Errorf("could not read cache file '%s': %w", cachePath, err)
	}

	err = json.Unmarshal(content, &cc.cache)
	if err != nil {
		return nil, errors.Errorf("could not unmarshal cache file '%s': %w", cachePath, err)
	}
	return &cc, nil
}

func (cc *CredentialsCache) Write() error {
	log.Debug("writing cache to disk", "filename", cc.cachePath, "cache", cc)
	data, err := json.Marshal(cc.cache)
	if err != nil {
		return errors.Errorf("could not marshal cache: %w", err)
	}
	err = os.WriteFile(cc.cachePath, data, 0600)
	if err != nil {
		return errors.Errorf("could not write cache to disk: %w", err)
	}
	return nil
}

func (cc *CredentialsCache) SetCredentials(profileName string, region string, credentials *types.Credentials) {
	log.Debug("set cache", "profileName", profileName, "AccessKeyId", credentials.AccessKeyId)
	cc.cache.ProfileCache[profileName] = &CachedProfile{
		AccessKeyId:     credentials.AccessKeyId,
		SecretAccessKey: credentials.SecretAccessKey,
		SessionToken:    credentials.SessionToken,
		Expiration:      credentials.Expiration,
		Region:          &region,
	}
}

func (cc *CredentialsCache) HasProfile(profileName string) bool {
	profile, ok := cc.cache.ProfileCache[profileName]
	if !ok {
		return false
	}
	if time.Now().After(*profile.Expiration) {
		log.Debug("cache for the profile has expired", "profileName", profileName)
		delete(cc.cache.ProfileCache, profileName)
		return false
	}
	return true
}

func (cc *CredentialsCache) GetProfile(profileName string) *CachedProfile {
	if !cc.HasProfile(profileName) {
		return nil
	}
	return cc.cache.ProfileCache[profileName]
}
