package cache

import (
	"os"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/k0kubun/pp"
)

func createTestCacheFileValid(t *testing.T, tmpPath string) error {
	t.Helper()
	cache := `
	{
		"profile_cache": {
			"test-profile": {
				"access_key_id": "test id",
				"secret_access_key": "test secret key",
				"session_token": "test session token",
				"expiration": "2006-01-02T15:04:05Z",
				"region": "test-region"
			}
		},
		"sso_cache": {
			"test": {
				"access_token": "test token",
				"expires_at": "2006-01-02T15:04:05Z",
				"refresh_token": "test refresh token",
				"sso_region": "test region"
			}
		}
	}
	`
	return os.WriteFile(tmpPath, []byte(cache), 0600)
}

func cat(t *testing.T, path string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	t.Logf("content: %s", content)
}

func createTestCacheFileInvalid(t *testing.T, tmpPath string) error {
	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte("this is not valid json"))
	file.Close()

	cat(t, tmpPath)

	return err
}

func TestNewCredentialsCacheExists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmpDir)

	xdg.Reload()

	tmpPath := tmpDir + "/" + GOSUME_CACHE_FILENAME
	err := createTestCacheFileValid(t, tmpPath)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	cc, err := Load()
	if err != nil {
		t.Errorf("oh no: %s", err)
	}
	pp.Print(cc)

	if cc.cachePath != tmpPath {
		t.Errorf("expected cache path to be %s, got %s", tmpPath, cc.cachePath)
	}

	const profile = "test-profile"
	if _, ok := cc.cache.ProfileCache[profile]; !ok {
		t.Errorf("expected cache to contain profile '%s'", profile)
	}

	const ssoSession = "test"
	if _, ok := cc.cache.SsoCache[ssoSession]; !ok {
		t.Errorf("expected cache to contain sso session '%s'", ssoSession)
	}
}

func TestNewCredentialsCacheNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmpDir)

	xdg.Reload()

	cache, err := Load()
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	tmpPath := tmpDir + "/" + GOSUME_CACHE_FILENAME
	if cache.cachePath != tmpPath {
		t.Errorf("expected cache path to be %s, got %s", tmpPath, cache.cachePath)
	}

	if len(cache.cache.ProfileCache) != 0 {
		t.Errorf("expected cache to be empty, got %d", len(cache.cache.ProfileCache))
	}

}

func TestNewCredentialsCacheNotReadable(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmpDir)

	xdg.Reload()

	tmpPath := tmpDir + "/" + GOSUME_CACHE_FILENAME
	err := createTestCacheFileValid(t, tmpPath)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	err = os.Chmod(tmpPath, 0200)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	cache, err := Load()
	if err == nil || cache != nil {
		t.Errorf("oh no: %s", err)
	}

	if !strings.HasPrefix(err.Error(), "could not read cache file") {
		t.Errorf("expected error to be about not being able to read the cache file")
	}

}

func TestNewCredentialsCacheNotJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmpDir)

	xdg.Reload()

	tmpPath := tmpDir + "/" + GOSUME_CACHE_FILENAME
	err := createTestCacheFileInvalid(t, tmpPath)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	cache, err := Load()
	if err == nil || cache != nil {
		t.Errorf("oh no: %s", err)
	}

	if !strings.HasPrefix(err.Error(), "could not unmarshal cache file") {
		t.Errorf("expected error to be about not being able to read the cache file")
	}
}
