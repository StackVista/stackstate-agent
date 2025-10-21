package state

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestDefaultCheckStateManagerConfig(t *testing.T) {
	csm := NewCheckStateManager()

	assert.Equal(t, "/opt/datadog-agent/run", csm.Config.StateRootPath)
	assert.Equal(t, 10*time.Minute, csm.Config.CacheExpirationDuration)
	assert.Equal(t, 10*time.Minute, csm.Config.CachePurgeDuration)
}

func TestReadFromDisk(t *testing.T) {
	csm := NewCheckStateManager()
	diskFileValue, err := csm.readFromDisk("this-key-does-not-exist")

	assert.Equal(t, "{}", diskFileValue)
	assert.Equal(t, nil, err)
}

func TestGetFileForKey(t *testing.T) {
	csm := NewCheckStateManager()
	filePath, err := csm.getFileForKey("this-key-does-not-exist")

	assert.Equal(t, fmt.Sprintf("%s/this-key-does-not-exist", csm.Config.StateRootPath), filePath)
	assert.Equal(t, nil, err)
}

func TestWriteToDisk(t *testing.T) {
	// Create a directory for the temp write to disk tests
	// We do not need to check for an error here, if this fails then the rest will
	var _ = os.Mkdir("temp-write-to-disk-test", os.ModePerm)

	os.Setenv("DD_CHECK_STATE_ROOT_PATH", "./temp-write-to-disk-test")

	// Create a file on the disk
	csm := NewCheckStateManager()
	err := csm.writeToDisk(fmt.Sprintf("temp-write-to-disk-test-%f", rand.Float64()), "random-data")

	assert.Equal(t, nil, err)
}

func TestCheckStateManagerConfigOverrides(t *testing.T) {
	os.Setenv("DD_CHECK_STATE_ROOT_PATH", "/my/custom/path")
	os.Setenv("DD_CHECK_STATE_EXPIRATION_DURATION", "5m")
	os.Setenv("DD_CHECK_STATE_PURGE_DURATION", "15m")

	csm := NewCheckStateManager()

	assert.Equal(t, "/my/custom/path", csm.Config.StateRootPath)
	assert.Equal(t, 5*time.Minute, csm.Config.CacheExpirationDuration)
	assert.Equal(t, 15*time.Minute, csm.Config.CachePurgeDuration)
}

func TestCheckStateManager_NonExistentGetState(t *testing.T) {
	os.Setenv("DD_CHECK_STATE_ROOT_PATH", "./file-does-not-exist")

	csm := NewCheckStateManager()

	stateKey := "random-state-key"
	expectedState := "{}"

	state, err := csm.GetState(stateKey)
	assert.NoError(t, err)
	assert.Equal(t, expectedState, state)

	// call GetState again, thereby fetching it from the cache
	state, err = csm.GetState(stateKey)
	assert.NoError(t, err, "unexpected error occurred when getting state for %s", stateKey)
	assert.Equal(t, expectedState, state)

	actualState, found := csm.Cache.Get(stateKey)
	assert.True(t, found, "%s not found in the check state manager cache", stateKey)
	assert.Equal(t, expectedState, actualState)
}

func TestCheckStateManager_GetState(t *testing.T) {
	os.Setenv("DD_CHECK_STATE_ROOT_PATH", "./testdata")

	csm := NewCheckStateManager()

	stateKey := "mycheck:state"
	expectedState := "{\"a\":\"b\"}"

	state, err := csm.GetState(stateKey)
	assert.NoError(t, err)
	assert.Equal(t, expectedState, state)

	// call GetState again, thereby fetching it from the cache
	state, err = csm.GetState(stateKey)
	assert.NoError(t, err, "unexpected error occurred when getting state for %s", stateKey)
	assert.Equal(t, expectedState, state)

	actualState, found := csm.Cache.Get(stateKey)
	assert.True(t, found, "%s not found in the check state manager cache", stateKey)
	assert.Equal(t, expectedState, actualState)
}

func TestCheckStateManager_SetStateWithUpdate(t *testing.T) {
	os.Setenv("DD_CHECK_STATE_ROOT_PATH", "./testdata")

	csm := NewCheckStateManager()

	stateKey := "setstatecheck:state"
	expectedState := "{\"c\":\"d\"}"

	err := csm.SetState(stateKey, expectedState)
	assert.NoError(t, err, "unexpected error occurred when setting state for %s", stateKey)

	state, err := csm.GetState(stateKey)
	assert.NoError(t, err, "unexpected error occurred when getting state for %s", stateKey)
	assert.Equal(t, expectedState, state)

	actualState, found := csm.Cache.Get(stateKey)
	assert.True(t, found, "%s not found in the check state manager cache", stateKey)
	assert.Equal(t, expectedState, actualState)

	updatedState := "{\"e\":\"f\"}"
	err = csm.SetState(stateKey, updatedState)
	assert.NoError(t, err, "unexpected error occurred when setting state for %s", stateKey)

	actualState, found = csm.Cache.Get(stateKey)
	assert.True(t, found, "%s not found in the check state manager cache", stateKey)
	assert.Equal(t, updatedState, actualState)

	actualState, found = csm.Cache.Get(stateKey)
	assert.True(t, found, "%s not found in the check state manager cache", stateKey)
	assert.Equal(t, updatedState, actualState)

	err = os.RemoveAll("./testdata/setstatecheck")
	assert.NoError(t, err, "error cleaning up test data file")
}
