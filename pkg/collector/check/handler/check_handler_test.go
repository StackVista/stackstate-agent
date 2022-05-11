package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/stretchr/testify/assert"
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
			checkHandler: NewCheckHandler(&check.TestCheck{Name: "my-check-handler-test-check"}, &check.TestCheckReloader{},
				integration.Data{1, 2, 3}, integration.Data{0, 0, 0}),
			expectedCHString:     "my-check-handler-test-check",
			expectedCHName:       "my-check-handler-test-check",
			expectedInitConfig:   integration.Data{0, 0, 0},
			expectedConfig:       integration.Data{1, 2, 3},
			expectedConfigSource: "test-config-source",
		},
		{
			testCase:             "no-check check handler",
			checkHandler:         MakeCheckNoHandler("no-check", &check.TestCheckReloader{}),
			expectedCHString:     "no-check-name",
			expectedCHName:       "no-check",
			expectedConfigSource: "no-source",
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

}

func TestCheckHandler_Transactions(t *testing.T) {
	testTxManager := transactionmanager.NewMockTransactionManager()

	ch := NewCheckHandler(&check.TestCheck{Name: "my-check-handler-test-check"}, &check.TestCheckReloader{},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*checkHandler)

	ch.Start()

	for _, tc := range []struct {
		testCase            string
		completeTransaction func()
	}{
		{
			testCase: "Transaction completed with transaction rollback",
			completeTransaction: func() {
				testTxManager.CurrentTransactionNotifyChannel <- transactionmanager.RollbackTransaction{}
			},
		},
		{
			testCase: "Transaction completed with transaction eviction",
			completeTransaction: func() {
				testTxManager.CurrentTransactionNotifyChannel <- transactionmanager.EvictedTransaction{}
			},
		},
		{
			testCase: "Transaction completed with transaction complete",
			completeTransaction: func() {
				testTxManager.CurrentTransactionNotifyChannel <- transactionmanager.CompleteTransaction{}
			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			ch.StartTransaction()

			time.Sleep(50 * time.Millisecond)
			transaction1 := ch.currentTransaction
			assert.Equal(t, transaction1, testTxManager.CurrentTransaction)

			// attempt to start new transaction before 1 has finished, this should be blocked
			ch.StartTransaction()

			// wait a bit and assert that we're still processing Transaction1 instead of the attempted Transaction2
			time.Sleep(50 * time.Millisecond)
			assert.Equal(t, transaction1, testTxManager.CurrentTransaction)

			// complete Transaction1
			tc.completeTransaction()
			time.Sleep(50 * time.Millisecond)

			transaction2 := ch.currentTransaction
			// assert that the transaction changed
			assert.NotEqual(t, transaction1, ch.currentTransaction)
			// wait a bit and assert that we've started processing Transaction2
			assert.Equal(t, transaction2, testTxManager.CurrentTransaction)

			// complete Transaction2
			testTxManager.CurrentTransactionNotifyChannel <- transactionmanager.CompleteTransaction{}
		})
	}

	time.Sleep(100 * time.Millisecond)
	ch.Stop()
}

func TestCheckHandler_Shutdown(t *testing.T) {
	testTxManager := transactionmanager.NewMockTransactionManager()
	ch := NewCheckHandler(&check.TestCheck{Name: "my-check-handler-test-check"}, &check.TestCheckReloader{},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*checkHandler)

	ch.Start()

	ch.StartTransaction()

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, ch.currentTransaction, testTxManager.CurrentTransaction)

	ch.Stop()

}
