package gosume

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/adrg/xdg"
	"github.com/aws/aws-sdk-go-v2/aws"
)

func createTestCacheFile(tmpPath string, validJSON bool) error {
	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	if validJSON {
		testProfile := &CachedProfile{
			AccessKeyId:     aws.String("test-access-key-id"),
			SecretAccessKey: aws.String("test-secret-access-key"),
			SessionToken:    aws.String("test-session-token"),
			Expiration:      aws.Time(time.Now().Add(1 * time.Hour)),
			Region:          aws.String("test-region"),
		}

		json, err := json.Marshal(map[string]*CachedProfile{"test-profile": testProfile})
		if err != nil {
			return err
		}

		file.Write(json)
	} else {
		file.Write([]byte("this is not valid json"))
	}

	return file.Close()
}

func TestNewCredentialsCacheExists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmpDir)

	xdg.Reload()

	tmpPath := tmpDir + "/" + GOSUME_CACHE_FILENAME
	err := createTestCacheFile(tmpPath, true)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	cache, err := NewCredentialsCache()
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	if cache.cachePath != tmpPath {
		t.Errorf("expected cache path to be %s, got %s", tmpPath, cache.cachePath)
	}

	if _, ok := cache.cache["test-profile"]; !ok {
		t.Errorf("expected cache to contain test-profile")
	}
}

func TestNewCredentialsCacheNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmpDir)

	xdg.Reload()

	cache, err := NewCredentialsCache()
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	tmpPath := tmpDir + "/" + GOSUME_CACHE_FILENAME
	if cache.cachePath != tmpPath {
		t.Errorf("expected cache path to be %s, got %s", tmpPath, cache.cachePath)
	}

	if len(cache.cache) != 0 {
		t.Errorf("expected cache to be empty, got %d", len(cache.cache))
	}

}

func TestNewCredentialsCacheNotReadable(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmpDir)

	xdg.Reload()

	tmpPath := tmpDir + "/" + GOSUME_CACHE_FILENAME
	err := createTestCacheFile(tmpPath, true)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	err = os.Chmod(tmpPath, 0200)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	cache, err := NewCredentialsCache()
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
	err := createTestCacheFile(tmpPath, false)
	if err != nil {
		t.Errorf("oh no: %s", err)
	}

	cache, err := NewCredentialsCache()
	if err == nil || cache != nil {
		t.Errorf("oh no: %s", err)
	}

	if !strings.HasPrefix(err.Error(), "could not unmarshal cache file") {
		t.Errorf("expected error to be about not being able to read the cache file")
	}
}
