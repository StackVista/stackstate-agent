package transactionmanager

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTransactionManager_HappyFlow(t *testing.T) {
	txManager := newTransactionManager(100, 250*time.Millisecond, 500*time.Millisecond,
		500*time.Millisecond).(*transactionManager)

	// assert that we're starting on a clean slate
	assert.Equal(t, txManager.TransactionCount(), 0)

	// start a transaction and assert it
	txID := uuid.New().String()
	txNotifyChannel := make(chan interface{})
	txManager.StartTransaction("checkID", txID, txNotifyChannel)
	assertTransaction(t, txManager, txID, InProgress, map[string]*Action{})

	// commit 15 actions and assert them
	actions := make(map[string]*Action, 5)
	for i := 0; i < 15; i++ {
		actID := fmt.Sprintf("action-%d", i)
		commitAssertAction(t, txManager, txID, actID, actions)
	}

	// acknowledge 15 action and assert them
	for i := 0; i < 15; i++ {
		actID := fmt.Sprintf("action-%d", i)
		txManager.AcknowledgeAction(txID, actID)
		time.Sleep(50 * time.Millisecond) // give the transaction manager a bit of time to acknowledge the action before asserting
		actions[actID] = &Action{ActionID: actID, Acknowledged: true}
		assertTransaction(t, txManager, txID, InProgress, actions)
	}

	// start a transaction and assert it
	txManager.CompleteTransaction(txID)

	select {
	case completeMsg := <-txNotifyChannel:
		assert.Equal(t, CompleteTransaction{TransactionID: txID}, completeMsg)
	case <-time.After(1 * time.Second):
		t.Fail()
	}

	assertTransaction(t, txManager, txID, Succeeded, actions)

	defer txManager.Stop()

	// sleep and wait for automatic cleanup to remove the successful transaction
	time.Sleep(1 * time.Second)
	assert.Equal(t, txManager.TransactionCount(), 0)
}

func TestTransactionManager_SetState(t *testing.T) {
	txManager := newTransactionManager(100, 250*time.Millisecond, 500*time.Millisecond,
		500*time.Millisecond).(*transactionManager)

	// assert that we're starting on a clean slate
	assert.Equal(t, txManager.TransactionCount(), 0)

	// start a transaction and assert it
	txID := uuid.New().String()
	txNotifyChannel := make(chan interface{})
	txManager.StartTransaction("checkID", txID, txNotifyChannel)
	assertTransaction(t, txManager, txID, InProgress, map[string]*Action{})

	txManager.SetState(txID, "my-test-state", "{\"a\": \"b\", \"c\": 4, \"d\": [1, 2, 3]}")

	// start a transaction and assert it
	txManager.CompleteTransaction(txID)

	select {
	case completeMsg := <-txNotifyChannel:
		expectedComplete := CompleteTransaction{
			TransactionID: txID,
			State: &TransactionState{
				Key:   "my-test-state",
				State: "{\"a\": \"b\", \"c\": 4, \"d\": [1, 2, 3]}",
			},
		}
		assert.Equal(t, expectedComplete, completeMsg)
	case <-time.After(1 * time.Second):
		t.Fail()
	}

	defer txManager.Stop()

	// sleep and wait for automatic cleanup to remove the successful transaction
	time.Sleep(1 * time.Second)
	assert.Equal(t, txManager.TransactionCount(), 0)
}

func TestTransactionManager_TransactionRollback(t *testing.T) {
	txManager := newTransactionManager(100, 100*time.Millisecond, 1*time.Second,
		1*time.Second).(*transactionManager)

	txNotifyChannel := make(chan interface{})

	for _, tc := range []struct {
		testCase       string
		operation      func(txID string, t *testing.T, manager *transactionManager) map[string]*Action
		rollbackReason string
	}{
		{
			testCase: "Transaction rollback triggered by external party (check handler)",
			operation: func(txID string, t *testing.T, manager *transactionManager) (actions map[string]*Action) {
				txManager.RollbackTransaction(txID, "check failed")
				return
			},
			rollbackReason: "check failed",
		},
		{
			testCase: "Transaction rollback triggered by an un-acknowledged action",
			operation: func(txID string, t *testing.T, manager *transactionManager) map[string]*Action {
				actions := make(map[string]*Action, 1)
				actID := uuid.New().String()
				commitAssertAction(t, txManager, txID, actID, actions)
				txManager.CompleteTransaction(txID)
				return actions
			},
			rollbackReason: "Not all actions have been acknowledged, rolling back transaction",
		},
		{
			testCase: "Transaction rollback triggered by rejected action",
			operation: func(txID string, t *testing.T, manager *transactionManager) map[string]*Action {
				actions := make(map[string]*Action, 1)
				actID := uuid.New().String()
				commitAssertAction(t, txManager, txID, actID, actions)

				txManager.RejectAction(txID, actID, "forced rejection")

				actions[actID].Acknowledged = true

				return actions
			},
			rollbackReason: "rejected action",
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			txID := uuid.New().String()
			txManager.StartTransaction("checkID", txID, txNotifyChannel)

			assertTransaction(t, txManager, txID, InProgress, map[string]*Action{})

			actions := tc.operation(txID, t, txManager)

			assertTransaction(t, txManager, txID, Failed, actions)

			completeMsg := <-txNotifyChannel
			rollbackTransaction := completeMsg.(RollbackTransaction)
			assert.Equal(t, txID, rollbackTransaction.TransactionID)
			assert.Contains(t, rollbackTransaction.Reason, tc.rollbackReason)

		})
	}

	// sleep and wait for automatic cleanup to remove the successful transaction
	time.Sleep(1 * time.Second)
	assert.Equal(t, txManager.TransactionCount(), 0)

	close(txNotifyChannel)
	txManager.Stop()
}

func TestTransactionManager_TransactionTimeout(t *testing.T) {
	staleTimeout := 100 * time.Millisecond
	txManager := newTransactionManager(100, 10*time.Millisecond, staleTimeout,
		750*time.Millisecond).(*transactionManager)

	txNotifyChannel := make(chan interface{})

	for _, tc := range []struct {
		testCase  string
		operation func(txID string, t *testing.T, manager *transactionManager) map[string]*Action
	}{
		{
			testCase: "Transaction timeout after becoming stale (no actions)",
			operation: func(txID string, t *testing.T, manager *transactionManager) (actions map[string]*Action) {
				return
			},
		},
		{
			testCase: "Transaction timeout after becoming stale for a second time with actions",
			operation: func(txID string, t *testing.T, manager *transactionManager) map[string]*Action {
				// assert that we have a InProgress transaction
				actions := make(map[string]*Action, 0)
				assertTransaction(t, txManager, txID, InProgress, actions)

				// sleep for staleTimeout and assert that the transaction has become stale
				time.Sleep(staleTimeout)
				assertTransaction(t, txManager, txID, Stale, actions)

				// commit an action and assert that the transaction is again in progress
				actID := uuid.New().String()
				commitAssertAction(t, txManager, txID, actID, actions)

				assertTransaction(t, txManager, txID, InProgress, actions)

				return actions
			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			txID := uuid.New().String()
			txManager.StartTransaction("CheckID", txID, txNotifyChannel)
			assertTransaction(t, txManager, txID, InProgress, map[string]*Action{})

			actions := tc.operation(txID, t, txManager)

			time.Sleep(staleTimeout)

			assertTransaction(t, txManager, txID, Stale, actions)

			// wait for the eviction notification
			notify := <-txNotifyChannel
			assert.Equal(t, EvictedTransaction{TransactionID: txID}, notify)
		})
	}

	assert.Equal(t, txManager.TransactionCount(), 0)

	close(txNotifyChannel)
	txManager.Stop()
}

func TestTransactionManager_ErrorHandling(t *testing.T) {
	txManager := newTransactionManager(100, 100*time.Millisecond, 1*time.Second,
		1*time.Second).(*transactionManager)

	txNotifyChannel := make(chan interface{})

	for _, tc := range []struct {
		testCase  string
		operation func(t *testing.T, manager *transactionManager)
	}{
		{
			testCase: "Commit action for a non-existing transaction",
			operation: func(t *testing.T, manager *transactionManager) {
				txID := uuid.New().String()
				actID := uuid.New().String()
				txManager.CommitAction(txID, actID)

				// assert that we don't have a transaction for txID and no action for actID
				_, notFoundError := txManager.GetTransaction(txID)
				assert.Error(t, notFoundError, "Transaction %s not found in the transaction map", txID)
				txManager.mux.RLock()
				for _, tx := range txManager.transactions {
					_, found := tx.Actions[actID]
					assert.False(t, found)
				}
				txManager.mux.RUnlock()
			},
		},
		{
			testCase: "Acknowledge a non-existing action for a transaction",
			operation: func(t *testing.T, manager *transactionManager) {
				txID := uuid.New().String()
				actions := make(map[string]*Action, 0)

				txManager.StartTransaction("checkID", txID, txNotifyChannel)
				assertTransaction(t, txManager, txID, InProgress, actions)

				actID := uuid.New().String()
				commitAssertAction(t, txManager, txID, actID, actions)

				txManager.AcknowledgeAction(txID, "non-existing-action")
				assertTransaction(t, txManager, txID, InProgress, actions)

				completeMsg := <-txNotifyChannel
				assert.Equal(t, EvictedTransaction{TransactionID: txID}, completeMsg)

			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			tc.operation(t, txManager)
		})
	}

	defer txManager.Stop()
	assert.Equal(t, txManager.TransactionCount(), 0)
}

func assertTransaction(t *testing.T, txManager *transactionManager, txID string, state TransactionStatus,
	actions map[string]*Action) {
	// give the transaction checkmanager a bit of time to insert the transaction before running the assertion
	time.Sleep(20 * time.Millisecond)
	txManager.mux.RLock()
	transaction, found := txManager.transactions[txID]
	assert.True(t, found, "Transaction %s not found in the transaction map", txID)
	assert.Equal(t, txID, transaction.TransactionID)
	assert.Equal(t, state, transaction.Status)
	assert.Equal(t, len(actions), len(transaction.Actions))
	for _, action := range transaction.Actions {
		expectedAction, found := actions[action.ActionID]
		assert.True(t, found)
		assert.Equal(t, expectedAction.ActionID, action.ActionID)
		assert.Equal(t, expectedAction.Acknowledged, action.Acknowledged)
	}
	txManager.mux.RUnlock()
}

func commitAssertAction(t *testing.T, txManager *transactionManager, txID, actID string, actions map[string]*Action) {
	txManager.CommitAction(txID, actID)
	actions[actID] = &Action{ActionID: actID, Acknowledged: false}
	assertTransaction(t, txManager, txID, InProgress, actions)
}
