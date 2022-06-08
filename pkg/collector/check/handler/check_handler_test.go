package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/state"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
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
			testCase: "my-check-handler-test-check check handler",
			checkHandler: NewCheckHandler(&check.STSTestCheck{Name: "my-check-handler-test-check"}, &check.TestCheckReloader{},
				integration.Data{1, 2, 3}, integration.Data{0, 0, 0}),
			expectedCHString:     "my-check-handler-test-check",
			expectedCHName:       "my-check-handler-test-check",
			expectedInitConfig:   integration.Data{0, 0, 0},
			expectedConfig:       integration.Data{1, 2, 3},
			expectedConfigSource: "test-config-source",
		},
		{
			testCase:             "no-check check handler",
			checkHandler:         MakeNonTransactionalCheckHandler(NewCheckIdentifier("no-check"), &check.TestCheckReloader{}),
			expectedCHString:     "no-check",
			expectedCHName:       "no-check",
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

			cr := tc.checkHandler.GetCheckReloader().(*check.TestCheckReloader)
			assert.Equal(t, 0, cr.Reloaded)
			tc.checkHandler.Reload()
			assert.Equal(t, 1, cr.Reloaded)
			tc.checkHandler.Reload()
			assert.Equal(t, 2, cr.Reloaded)
		})
	}

	// stop the transactional components
	transactionbatcher.GetTransactionalBatcher().Stop()

}

func TestCheckHandler_Transactions(t *testing.T) {
	testTxManager := transactionmanager.NewMockTransactionManager()
	transactionbatcher.NewMockTransactionalBatcher()

	ch := NewCheckHandler(&check.STSTestCheck{Name: "my-check-handler-test-check"}, &check.TestCheckReloader{},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*checkHandler)

	for _, tc := range []struct {
		testCase            string
		completeTransaction func()
	}{
		{
			testCase: "Transaction completed with transaction rollback",
			completeTransaction: func() {
				testTxManager.GetCurrentTransactionNotifyChannel() <- transactionmanager.RollbackTransaction{}
			},
		},
		{
			testCase: "Transaction completed with transaction eviction",
			completeTransaction: func() {
				testTxManager.GetCurrentTransactionNotifyChannel() <- transactionmanager.EvictedTransaction{}
			},
		},
		{
			testCase: "Transaction completed with transaction complete",
			completeTransaction: func() {
				testTxManager.GetCurrentTransactionNotifyChannel() <- transactionmanager.CompleteTransaction{}
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
			tc.completeTransaction()
			time.Sleep(50 * time.Millisecond)

			// wait a bit and assert that we've started processing Transaction2
			assert.Equal(t, transaction2, testTxManager.GetCurrentTransaction())

			// complete Transaction2
			testTxManager.GetCurrentTransactionNotifyChannel() <- transactionmanager.CompleteTransaction{}
		})
	}

	time.Sleep(100 * time.Millisecond)
	ch.Stop()

	// stop the transactional components
	transactionbatcher.GetTransactionalBatcher().Stop()
	transactionmanager.GetTransactionManager().Stop()
}

func TestCheckHandler_State(t *testing.T) {
	os.Setenv("DD_CHECK_STATE_ROOT_PATH", "./testdata")
	state.InitCheckStateManager()

	ch := NewCheckHandler(&check.STSTestCheck{Name: "my-check-handler-test-check"}, &check.TestCheckReloader{},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*checkHandler)

	stateKey := "my-check-handler-test-check:state"

	actualState := ch.GetState(stateKey)
	expectedState := "{\"a\":\"b\"}"
	assert.Equal(t, expectedState, actualState)

	checkState, err := state.GetCheckStateManager().GetState(stateKey)
	assert.NoError(t, err, "unexpected error occurred when trying to get state for", stateKey)
	assert.Equal(t, expectedState, checkState)

	updatedState := "{\"e\":\"f\"}"
	err = ch.SetState(stateKey, updatedState)
	assert.NoError(t, err, "unexpected error occurred when setting state for %s", stateKey)

	checkState, err = state.GetCheckStateManager().GetState(stateKey)
	assert.NoError(t, err, "unexpected error occurred when trying to get state for", stateKey)
	assert.Equal(t, updatedState, checkState)

	// reset to original
	err = ch.SetState(stateKey, expectedState)
	assert.NoError(t, err, "unexpected error occurred when setting state for %s", stateKey)
}

// Reset state to original, kept as a separate test in case of a test failure in TestCheckHandler_State
func TestCheckHandler_Reset_State(t *testing.T) {
	os.Setenv("DD_CHECK_STATE_ROOT_PATH", "./testdata")
	state.InitCheckStateManager()

	ch := NewCheckHandler(&check.STSTestCheck{Name: "my-check-handler-test-check"}, &check.TestCheckReloader{},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*checkHandler)

	stateKey := "my-check-handler-test-check:state"
	expectedState := "{\"a\":\"b\"}"

	// reset state to original
	err := ch.SetState(stateKey, expectedState)
	assert.NoError(t, err, "unexpected error occurred when setting state for %s", stateKey)

	checkState, err := state.GetCheckStateManager().GetState(stateKey)
	assert.NoError(t, err, "unexpected error occurred when trying to get state for", stateKey)
	assert.Equal(t, expectedState, checkState)
}

func TestCheckHandler_Shutdown(t *testing.T) {
	testTxManager := transactionmanager.NewMockTransactionManager()
	ch := NewCheckHandler(&check.STSTestCheck{Name: "my-check-handler-test-check"}, &check.TestCheckReloader{},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*checkHandler)

	transactionID := ch.StartTransaction()

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, transactionID, testTxManager.GetCurrentTransaction())

	ch.Stop()

	// stop the transactional components
	transactionmanager.GetTransactionManager().Stop()

}
