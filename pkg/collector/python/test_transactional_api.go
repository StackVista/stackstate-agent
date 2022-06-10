//go:build python && test
// +build python,test

package python

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/checkmanager"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// #include <datadog_agent_rtloader.h>
import "C"

func testStartTransaction(t *testing.T) {
	SetupTransactionalComponents()
	mockTransactionalManager := transactionmanager.GetTransactionManager().(*transactionmanager.MockTransactionManager)

	testCheck := &check.STSTestCheck{Name: "check-id-start-transaction"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})
	checkId := C.CString(testCheck.String())

	StartTransaction(checkId)
	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	transactionID := mockTransactionalManager.GetCurrentTransaction()
	assert.NotEmpty(t, transactionID)

	checkmanager.GetCheckManager().UnsubscribeCheckHandler(testCheck.ID())
}

func testStopTransaction(t *testing.T) {
	SetupTransactionalComponents()
	mockTransactionalBatcher := transactionbatcher.GetTransactionalBatcher().(*transactionbatcher.MockTransactionalBatcher)
	mockTransactionalManager := transactionmanager.GetTransactionManager().(*transactionmanager.MockTransactionManager)

	testCheck := &check.STSTestCheck{Name: "check-id-stop-transaction"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})
	checkId := C.CString(testCheck.String())

	StartTransaction(checkId)
	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	transactionID := mockTransactionalManager.GetCurrentTransaction()

	StopTransaction(checkId)

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

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

	SetupTransactionalComponents()
	testCheck := &check.STSTestCheck{Name: "check-id-set-transaction-state"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	mockTransactionalManager := transactionmanager.GetTransactionManager().(*transactionmanager.MockTransactionManager)

	stateKeyString := "key"
	stateValueString := "state value"
	checkId := C.CString(testCheck.String())
	stateKey := C.CString(stateKeyString)
	stateValue := C.CString(stateValueString)

	StartTransaction(checkId)
	SetTransactionState(checkId, stateKey, stateValue)

	transactionID := mockTransactionalManager.GetCurrentTransaction()
	assert.NotEmpty(t, transactionID)

	expectedState := &transactionmanager.TransactionState{
		Key:   stateKeyString,
		State: stateValueString,
	}
	actualState := mockTransactionalManager.GetCurrentTransactionState()
	assert.Equal(t, expectedState, actualState)

	checkmanager.GetCheckManager().UnsubscribeCheckHandler(testCheck.ID())
}
