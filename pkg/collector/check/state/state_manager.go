package state

import (
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Invalid characters to clean up
	invalidChars = regexp.MustCompile("[^a-zA-Z0-9_-]")
)

// CheckStateAPI contains all the functions for setting and getting state from disk / memory
type CheckStateAPI interface {
	GetState(key string) (string, error)
	SetState(key, value string) error
	Clear()
}

// CheckStateManager is the default implementation for the CheckStateAPI that read / writes state from / to disk
type CheckStateManager struct {
	Config Config
	Cache  *cache.Cache
}

// NewCheckStateManager returns a pointer to an instance of CheckStateManager which implements the CheckStateAPI
func NewCheckStateManager() *CheckStateManager {
	config := GetStateConfig()
	return &CheckStateManager{
		Config: config,
		Cache:  cache.New(config.CacheExpirationDuration, config.CachePurgeDuration),
	}
}

// Return a file where to store the data. We split the key by ":", using the
// first prefix as directory, if present. This is useful for integrations, which
// use the check_id formed with $check_name:$hash
func (cs *CheckStateManager) getFileForKey(key string) (string, error) {
	paths := strings.SplitN(key, ":", 2)
	cleanedPath := invalidChars.ReplaceAllString(paths[0], "")
	if len(paths) == 1 {
		// If there is no colon, just return the key
		return filepath.Join(cs.Config.StateRootPath, cleanedPath), nil
	}
	// Otherwise, create the directory with a prefix
	err := os.MkdirAll(filepath.Join(cs.Config.StateRootPath, cleanedPath), 0700)
	if err != nil {
		return "", err
	}
	cleanedFile := invalidChars.ReplaceAllString(paths[1], "")
	return filepath.Join(cs.Config.StateRootPath, cleanedPath, cleanedFile), nil
}

// SetState stores data on disk in the config.StateRootPath directory
func (cs *CheckStateManager) SetState(key, value string) error {
	err := cs.writeToDisk(key, value)
	if err != nil {
		_ = log.Errorf("Unable to set a new state to disk. Error: %v", err)
		return err
	}
	// Insert / Update this in the CheckStateManager Cache
	cs.Cache.Set(key, value, cache.DefaultExpiration)

	return nil
}

func (cs *CheckStateManager) writeToDisk(key, value string) error {
	path, err := cs.getFileForKey(key)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, []byte(value), 0600)
	if err != nil {
		return err
	}

	return nil
}

// GetState returns a value previously stored, or an error that occurred when trying to retrieve the state for a given
// key
func (cs *CheckStateManager) GetState(key string) (string, error) {
	// see if we have this key in the cache, otherwise read it from Disk
	if value, found := cs.Cache.Get(key); found {
		if typeAssertionValue, ok := value.(string); ok {
			return typeAssertionValue, nil
		}
	}
	state, err := cs.readFromDisk(key)
	if err != nil {
		_ = log.Errorf("Error occurred loading state from disk. Error: %v", err)
		return "{}", err
	}
	// update the cache
	cs.Cache.Set(key, state, cache.DefaultExpiration)
	return state, nil
}

func (cs *CheckStateManager) readFromDisk(key string) (string, error) {
	path, err := cs.getFileForKey(key)
	if err != nil {
		return "{}", err
	}
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		return "{}", nil
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "{}", err
	}
	// return the string, removing any trailing new lines
	return strings.TrimSuffix(string(content), "\n"), nil
}

// Clear removes all the elements in the CheckStateManager Cache
func (cs *CheckStateManager) Clear() {
	cs.Cache.Flush()
}
