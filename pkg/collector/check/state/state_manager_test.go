package state

import (
	"github.com/stretchr/testify/assert"
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

func TestCheckStateManagerConfigOverrides(t *testing.T) {
	os.Setenv("DD_CHECK_STATE_ROOT_PATH", "/my/custom/path")
	os.Setenv("DD_CHECK_STATE_EXPIRATION_DURATION", "5m")
	os.Setenv("DD_CHECK_STATE_PURGE_DURATION", "15m")

	csm := NewCheckStateManager()

	assert.Equal(t, "/my/custom/path", csm.Config.StateRootPath)
	assert.Equal(t, 5*time.Minute, csm.Config.CacheExpirationDuration)
	assert.Equal(t, 15*time.Minute, csm.Config.CachePurgeDuration)
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
