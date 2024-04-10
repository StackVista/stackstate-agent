package handler

import (
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	"github.com/DataDog/datadog-agent/pkg/collector/check/state"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional/transactionmanager"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestCheckHandler(t *testing.T) {

	// init global transactionbatcher used by the check no handler
	transactionbatcher.NewMockTransactionalBatcher()

	for _, tc := range []struct {
		testCase             string
		checkHandler         CheckHandler
		expectedCHString     string
		expectedCHName       string
		expectedInitConfig   integration.Data
		expectedConfig       integration.Data
		expectedConfigSource string
	}{
		{
			testCase: "my-check-handler-test-check transactional check handler",
			checkHandler: NewTransactionalCheckHandler(&check.STSTestCheck{Name: "my-check-handler-test-check"},
				integration.Data{1, 2, 3}, integration.Data{0, 0, 0}),
			expectedCHString:     "my-check-handler-test-check",
			expectedCHName:       "my-check-handler-test-check",
			expectedInitConfig:   integration.Data{0, 0, 0},
			expectedConfig:       integration.Data{1, 2, 3},
			expectedConfigSource: "test-config-source",
		},
		{
			testCase: "no-check check handler",
			checkHandler: MakeNonTransactionalCheckHandler(NewCheckIdentifier("my-check-handler-test-check-2"),
				integration.Data{1, 2, 3}, integration.Data{0, 0, 0}),
			expectedCHString:     "my-check-handler-test-check-2",
			expectedCHName:       "my-check-handler-test-check-2",
			expectedInitConfig:   integration.Data{0, 0, 0},
			expectedConfig:       integration.Data{1, 2, 3},
			expectedConfigSource: "",
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			assert.Equal(t, tc.expectedCHString, tc.checkHandler.String())
			assert.Equal(t, check.ID(tc.expectedCHName), tc.checkHandler.ID())
			actualInstanceCfg, actualInitCfg := tc.checkHandler.GetConfig()
			assert.EqualValues(t, tc.expectedConfig, actualInstanceCfg)
			assert.EqualValues(t, tc.expectedInitConfig, actualInitCfg)
			assert.Equal(t, tc.expectedConfigSource, tc.checkHandler.ConfigSource())
		})
	}

	// stop the transactional components
	transactionbatcher.GetTransactionalBatcher().Stop()

}

func TestCheckHandler_Transactions(t *testing.T) {
	testTxManager := transactionmanager.NewMockTransactionManager()
	transactionbatcher.NewMockTransactionalBatcher()

	ch := NewTransactionalCheckHandler(&check.STSTestCheck{Name: "my-check-handler-transactions-test-check"},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*TransactionalCheckHandler)

	for _, tc := range []struct {
		testCase            string
		completeTransaction func(transaction string)
	}{
		{
			testCase: "Transaction completed with transaction rollback",
			completeTransaction: func(transaction string) {
				testTxManager.GetCurrentTransactionNotifyChannel() <- transactionmanager.DiscardTransaction{
					TransactionID: transaction, Reason: "",
				}
			},
		},
		{
			testCase: "Transaction completed with transaction eviction",
			completeTransaction: func(transaction string) {
				testTxManager.GetCurrentTransactionNotifyChannel() <- transactionmanager.EvictedTransaction{TransactionID: transaction}
			},
		},
		{
			testCase: "Transaction completed with transaction complete",
			completeTransaction: func(transaction string) {
				testTxManager.GetCurrentTransactionNotifyChannel() <- transactionmanager.CompleteTransaction{TransactionID: transaction}
			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			transaction1 := ch.StartTransaction()

			time.Sleep(50 * time.Millisecond)
			assert.Equal(t, transaction1, testTxManager.GetCurrentTransaction())

			// attempt to start new transaction before 1 has finished, this should be blocked
			transaction2 := ch.StartTransaction()

			// wait a bit and assert that we're still processing Transaction1 instead of the attempted Transaction2
			time.Sleep(50 * time.Millisecond)
			assert.Equal(t, transaction1, testTxManager.GetCurrentTransaction())

			// complete Transaction1
			tc.completeTransaction(transaction1)
			time.Sleep(50 * time.Millisecond)

			// wait a bit and assert that we've started processing Transaction2
			assert.Equal(t, transaction2, testTxManager.GetCurrentTransaction())

			// complete Transaction2
			testTxManager.GetCurrentTransactionNotifyChannel() <- transactionmanager.CompleteTransaction{TransactionID: transaction2}
		})
	}

	time.Sleep(100 * time.Millisecond)
	ch.Stop()
}

func TestCheckHandler_TransactionsIncorrectComplete(t *testing.T) {
	testTxManager := transactionmanager.NewMockTransactionManager()
	transactionbatcher.NewMockTransactionalBatcher()

	ch := NewTransactionalCheckHandler(&check.STSTestCheck{Name: "my-check-handler-transactions-incorrect-complete-test-check"},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*TransactionalCheckHandler)

	for _, tc := range []struct {
		testCase            string
		completeTransaction func(transaction string)
	}{
		{
			testCase: "Transaction completed with transaction rollback",
			completeTransaction: func(transaction string) {
				testTxManager.GetCurrentTransactionNotifyChannel() <- transactionmanager.DiscardTransaction{
					TransactionID: transaction, Reason: "",
				}
			},
		},
		{
			testCase: "Transaction completed with transaction eviction",
			completeTransaction: func(transaction string) {
				testTxManager.GetCurrentTransactionNotifyChannel() <- transactionmanager.EvictedTransaction{TransactionID: transaction}
			},
		},
		{
			testCase: "Transaction completed with transaction complete",
			completeTransaction: func(transaction string) {
				testTxManager.GetCurrentTransactionNotifyChannel() <- transactionmanager.CompleteTransaction{TransactionID: transaction}
			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			transaction1 := ch.StartTransaction()

			assert.Eventually(t, func() bool {
				return transaction1 == testTxManager.GetCurrentTransaction()
			}, 50*time.Millisecond, 10*time.Millisecond)

			// complete Transaction1
			incorrectTransactionID := "incorrect-transaction-id"
			tc.completeTransaction(incorrectTransactionID)

			assert.Eventually(t, func() bool {
				return transaction1 == testTxManager.GetCurrentTransaction()
			}, 50*time.Millisecond, 10*time.Millisecond)

			assert.Never(t, func() bool {
				return incorrectTransactionID == testTxManager.GetCurrentTransaction()
			}, 50*time.Millisecond, 10*time.Millisecond)

			transaction2 := ch.StartTransaction()
			assert.Never(t, func() bool {
				return transaction2 == testTxManager.GetCurrentTransaction()
			}, 50*time.Millisecond, 10*time.Millisecond)

			testTxManager.GetCurrentTransactionNotifyChannel() <- transactionmanager.CompleteTransaction{TransactionID: transaction1}
			testTxManager.GetCurrentTransactionNotifyChannel() <- transactionmanager.CompleteTransaction{TransactionID: transaction2}
		})
	}

	time.Sleep(100 * time.Millisecond)
	ch.Stop()
}

func TestCheckHandler_State(t *testing.T) {
	os.Setenv("DD_CHECK_STATE_ROOT_PATH", "./testdata")
	state.InitCheckStateManager()

	ch := NewTransactionalCheckHandler(&check.STSTestCheck{Name: "my-check-handler-state-test-check"},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*TransactionalCheckHandler)

	stateKey := "my-check-handler-test-check:state"

	actualState := ch.GetState(stateKey)
	expectedState := "{\"a\":\"b\"}"
	assert.Equal(t, expectedState, actualState)

	checkState, err := state.GetCheckStateManager().GetState(stateKey)
	assert.NoError(t, err, "unexpected error occurred when trying to get state for", stateKey)
	assert.Equal(t, expectedState, checkState)

	updatedState := "{\"e\":\"f\"}"
	ch.SetState(stateKey, updatedState)

	checkState, err = state.GetCheckStateManager().GetState(stateKey)
	assert.NoError(t, err, "unexpected error occurred when trying to get state for", stateKey)
	assert.Equal(t, updatedState, checkState)

	// reset to original
	ch.SetState(stateKey, expectedState)
}

// Reset state to original, kept as a separate test in case of a test failure in TestCheckHandler_State
func TestCheckHandler_Reset_State(t *testing.T) {
	os.Setenv("DD_CHECK_STATE_ROOT_PATH", "./testdata")
	state.InitCheckStateManager()

	ch := NewTransactionalCheckHandler(&check.STSTestCheck{Name: "my-check-handler-reset-state-test-check"},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*TransactionalCheckHandler)

	stateKey := "my-check-handler-test-check:state"
	expectedState := "{\"a\":\"b\"}"

	// reset state to original
	ch.SetState(stateKey, expectedState)

	checkState, err := state.GetCheckStateManager().GetState(stateKey)
	assert.NoError(t, err, "unexpected error occurred when trying to get state for", stateKey)
	assert.Equal(t, expectedState, checkState)
}

func TestCheckHandler_Shutdown(t *testing.T) {
	testTxManager := transactionmanager.NewMockTransactionManager()
	ch := NewTransactionalCheckHandler(&check.STSTestCheck{Name: "my-check-handler-shutdown-test-check"},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*TransactionalCheckHandler)

	transactionID := ch.StartTransaction()

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, transactionID, testTxManager.GetCurrentTransaction())

	ch.Stop()

}
