//go:build python && test
// +build python,test

package python

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/checkmanager"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

// #include <datadog_agent_rtloader.h>
import "C"

func testStartTransaction(t *testing.T) {
	testCheck := &check.STSTestCheck{Name: "check-id-start-transaction"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})
	checkId := C.CString(string(testCheck.ID()))

	StartTransaction(checkId)
	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	transactionID := mockTransactionManager.GetCurrentTransaction()
	assert.NotEmpty(t, transactionID)
}

func testStopTransaction(t *testing.T) {
	testCheck := &check.STSTestCheck{Name: "check-id-stop-transaction"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})
	checkId := C.CString(string(testCheck.ID()))

	StartTransaction(checkId)
	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	transactionID := mockTransactionManager.GetCurrentTransaction()

	StopTransaction(checkId)

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: &transactionbatcher.BatchTransaction{
			TransactionID:        transactionID,
			CompletedTransaction: true,
		},
		Health: map[string]health.Health{},
	}, actualTopology)

	checkmanager.GetCheckManager().UnsubscribeCheckHandler(testCheck.ID())
}

func testSetTransactionState(t *testing.T) {
	// Create a temp directory to store the state results in
	testDir, err := ioutil.TempDir("", "fake-datadog-run-")
	require.Nil(t, err, fmt.Sprintf("%v", err))
	defer os.RemoveAll(testDir)
	mockConfig := config.Mock()

	// Set the run path to the temp directory above, this will allow the persistent cache to have a folder to write into
	// Without doing the above persistent cache will generate a folder does not exist error
	mockConfig.Set("run_path", testDir)

	testCheck := &check.STSTestCheck{Name: "check-id-set-transaction-state"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	checkId := C.CString(string(testCheck.ID()))
	stateKey := C.CString("key")
	stateValue := C.CString("state value")

	SetTransactionState(checkId, stateKey, stateValue)

	retrievedStateValue := GetState(checkId, stateKey)

	assert.Equal(t, "state value", retrievedStateValue)
}
