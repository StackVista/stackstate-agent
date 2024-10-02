//go:build python && test

package python

import (
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check/handler"
	"github.com/DataDog/datadog-agent/pkg/collector/check/test"
	check2 "github.com/StackVista/stackstate-receiver-go-client/pkg/model/check"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/health"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionmanager"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// #include <datadog_agent_rtloader.h>
import "C"

func testStartTransaction(t *testing.T) {
	_, _, mockTransactionalManager, checkManager := handler.SetupMockTransactionalComponents()

	release := scopeInitCheckManager(checkManager)
	defer release()

	testCheck := &test.STSTestCheck{Name: "check-id-start-transaction"}
	checkManager.RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})
	checkId := C.CString(testCheck.String())

	StartTransaction(checkId)
	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	transactionID := mockTransactionalManager.GetCurrentTransaction()
	assert.NotEmpty(t, transactionID)

	checkManager.UnsubscribeCheckHandler(testCheck.ID())
}

func testDiscardTransaction(t *testing.T) {
	_, _, mockTransactionalManager, checkManager := handler.SetupMockTransactionalComponents()

	release := scopeInitCheckManager(checkManager)
	defer release()

	testCheck := &test.STSTestCheck{Name: "check-id-cancel-transaction"}
	checkManager.RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	checkId := C.CString(testCheck.String())
	cancelReasonString := "test-transacation-cancel"
	cancelReason := C.CString(cancelReasonString)

	StartTransaction(checkId)
	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	transactionID := mockTransactionalManager.GetCurrentTransaction()
	assert.NotEmpty(t, transactionID)

	// discard transaction
	DiscardTransaction(checkId, cancelReason)
	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	transactionID = mockTransactionalManager.GetCurrentTransaction()

	discardAction := mockTransactionalManager.NextAction().(transactionmanager.DiscardTransaction)
	assert.Equal(t, transactionmanager.DiscardTransaction{TransactionID: transactionID, Reason: cancelReasonString}, discardAction)

	checkManager.UnsubscribeCheckHandler(testCheck.ID())
}

func testStopTransaction(t *testing.T) {
	_, mockTransactionalBatcher, mockTransactionalManager, checkManager := handler.SetupMockTransactionalComponents()

	release := scopeInitCheckManager(checkManager)
	defer release()

	testCheck := &test.STSTestCheck{Name: "check-id-stop-transaction"}
	checkManager.RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})
	checkId := C.CString(testCheck.String())

	StartTransaction(checkId)
	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	transactionID := mockTransactionalManager.GetCurrentTransaction()

	StopTransaction(checkId)

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	actualTopology, found := mockTransactionalBatcher.GetCheckState(check2.CheckID(testCheck.ID()))
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: &transactionbatcher.BatchTransaction{
			TransactionID:        transactionID,
			CompletedTransaction: true,
		},
		Health: map[string]health.Health{},
	}, actualTopology)

	checkManager.UnsubscribeCheckHandler(testCheck.ID())
}

func testSetTransactionState(t *testing.T) {
	_, _, mockTransactionalManager, checkManager := handler.SetupMockTransactionalComponents()

	release := scopeInitCheckManager(checkManager)
	defer release()

	testCheck := &test.STSTestCheck{Name: "check-id-set-transaction-state"}
	checkManager.RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	stateKeyString := "key"
	stateValueString := "state value"
	checkId := C.CString(testCheck.String())
	stateKey := C.CString(stateKeyString)
	stateValue := C.CString(stateValueString)

	StartTransaction(checkId)
	SetTransactionState(checkId, stateKey, stateValue)

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	transactionID := mockTransactionalManager.GetCurrentTransaction()
	assert.NotEmpty(t, transactionID)

	expectedState := &transactionmanager.TransactionState{
		Key:   stateKeyString,
		State: stateValueString,
	}
	actualState := mockTransactionalManager.GetCurrentTransactionState()
	assert.Equal(t, expectedState, actualState)

	checkManager.UnsubscribeCheckHandler(testCheck.ID())
}
