//go:build python && test
// +build python,test

package python

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/checkmanager"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

// #include <datadog_agent_rtloader.h>
import "C"

func testSetAndGetState(t *testing.T) {
	// Create a temp directory to store the state results in
	testDir, err := ioutil.TempDir("", "fake-datadog-run-")
	require.Nil(t, err, fmt.Sprintf("%v", err))
	defer os.RemoveAll(testDir)
	mockConfig := config.Mock()

	// Set the run path to the temp directory above, this will allow the persistent cache to have a folder to write into
	// Without doing the above persistent cache will generate a folder does not exist error
	mockConfig.Set("run_path", testDir)

	SetupTransactionalComponents()
	testCheck := &check.STSTestCheck{Name: "check-id-set-state"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	checkId := C.CString(string(testCheck.ID()))
	stateKey := C.CString("state-id")
	stateBody := C.CString("state-body")

	SetState(checkId, stateKey, stateBody)
	retrievedStateBody := GetState(checkId, stateKey)

	assert.Equal(t, "state-body", retrievedStateBody)
}
