package gosume

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/adrg/xdg"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/charmbracelet/log"
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
	cache     map[string]*CachedProfile
	cachePath string
}

func NewCredentialsCache() (*CredentialsCache, error) {
	cachePath, err := xdg.CacheFile(GOSUME_CACHE_FILENAME)
	if err != nil {
		return nil, fmt.Errorf("could not determine cache location: %w", err)
	}
	log.Debug("found cache file", "filename", cachePath)

	cache := &CredentialsCache{
		cachePath: cachePath,
		cache:     make(map[string]*CachedProfile),
	}

	content, err := os.ReadFile(cachePath)
	if os.IsNotExist(err) {
		return cache, nil
	}

	if err != nil {
		return nil, fmt.Errorf("could not read cache file '%s': %w", cachePath, err)
	}

	err = json.Unmarshal(content, &cache.cache)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal cache file '%s': %w", cachePath, err)
	}
	return cache, nil
}

func (cc *CredentialsCache) writeCredentialsCacheToDisk() error {
	json, err := json.Marshal(cc.cache)
	if err != nil {
		return fmt.Errorf("could not marshal cache: %w", err)
	}
	err = os.WriteFile(cc.cachePath, json, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not write cache to disk: %w", err)
	}
	return nil
}

func (cc *CredentialsCache) setCredentials(profileName string, region string, credentials *types.Credentials) {
	log.Debug("set cache", "profileName", profileName, "AccessKeyId", credentials.AccessKeyId)
	cc.cache[profileName] = &CachedProfile{
		AccessKeyId:     credentials.AccessKeyId,
		SecretAccessKey: credentials.SecretAccessKey,
		SessionToken:    credentials.SessionToken,
		Expiration:      credentials.Expiration,
		Region:          &region,
	}
}

func (cc *CredentialsCache) HasProfile(profileName string) bool {
	profile, ok := cc.cache[profileName]
	if !ok {
		return false
	}
	if time.Now().After(*profile.Expiration) {
		log.Debug("cache for the profile has expired", "profileName", profileName)
		delete(cc.cache, profileName)
		return false
	}
	return true
}

func (cc *CredentialsCache) GetProfile(profileName string) *CachedProfile {
	if !cc.HasProfile(profileName) {
		return nil
	}
	return cc.cache[profileName]
}
