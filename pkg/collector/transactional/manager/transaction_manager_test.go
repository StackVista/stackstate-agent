package manager

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTransactionManager_HappyFlow(t *testing.T) {
	transactionManager := MakeTransactionManager(100, 100*time.Millisecond, 500*time.Millisecond,
		500*time.Millisecond)

	assert.False(t, transactionManager.running)
	transactionManager.Start()
	assert.True(t, transactionManager.running)

	// assert that we're starting on a clean slate
	assert.Len(t, transactionManager.Transactions, 0)

	completeFunc := func(transaction *IntakeTransaction) {
		// assert that LastUpdatedTimestamp was set in the last 50 milliseconds. i.e. just updated
		assert.True(t, transaction.LastUpdatedTimestamp.After(time.Now().Add(-50*time.Millisecond)))
		assert.Equal(t, Succeeded, transaction.State)
	}

	// start a transaction and assert it
	txID := uuid.New().String()
	transactionManager.TransactionChannel <- StartTransaction{TransactionID: txID, OnComplete: completeFunc}
	assertTransaction(t, transactionManager, txID, InProgress, map[string]*Action{})

	// commit 15 actions and assert them
	actions := make(map[string]*Action, 5)
	for i := 0; i < 15; i++ {
		actID := fmt.Sprintf("action-%d", i)
		commitAssertAction(t, transactionManager, txID, actID, actions)
	}
	// acknowledge 15 action and assert them
	for i := 0; i < 15; i++ {
		actID := fmt.Sprintf("action-%d", i)
		transactionManager.TransactionChannel <- AckAction{TransactionID: txID, ActionID: actID}
		actions[actID] = &Action{ActionID: actID, Acknowledged: true}
		assertTransaction(t, transactionManager, txID, InProgress, actions)
	}

	// start a transaction and assert it
	transactionManager.TransactionChannel <- CompleteTransaction{TransactionID: txID}
	assertTransaction(t, transactionManager, txID, Succeeded, actions)

	defer transactionManager.Stop()

	// sleep and wait for automatic cleanup to remove the successful transaction
	time.Sleep(1 * time.Second)
	assert.Len(t, transactionManager.Transactions, 0)
}

func TestTransactionManager_TransactionRollback(t *testing.T) {
	transactionManager := MakeTransactionManager(100, 100*time.Millisecond, 500*time.Millisecond,
		500*time.Millisecond)

	transactionManager.Start()

	for _, tc := range []struct {
		testCase  string
		operation func(txID string, t *testing.T, manager *TransactionManager) map[string]*Action
	}{
		{
			testCase: "Transaction rollback triggered by external party (check handler)",
			operation: func(txID string, t *testing.T, manager *TransactionManager) (actions map[string]*Action) {
				transactionManager.TransactionChannel <- RollbackTransaction{TransactionID: txID, Reason: "check failed"}
				return
			},
		},
		{
			testCase: "Transaction rollback triggered by an un-acknowledged action",
			operation: func(txID string, t *testing.T, manager *TransactionManager) map[string]*Action {
				actions := make(map[string]*Action, 1)
				actID := uuid.New().String()
				commitAssertAction(t, transactionManager, txID, actID, actions)

				transactionManager.TransactionChannel <- CompleteTransaction{TransactionID: txID}

				return actions
			},
		},
		{
			testCase: "Transaction rollback triggered by rejected action",
			operation: func(txID string, t *testing.T, manager *TransactionManager) map[string]*Action {
				actions := make(map[string]*Action, 1)
				actID := uuid.New().String()
				commitAssertAction(t, transactionManager, txID, actID, actions)

				transactionManager.TransactionChannel <- RejectAction{TransactionID: txID, ActionID: actID,
					Reason: "forced rejection"}

				actions[actID].Acknowledged = true

				return actions
			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			txID := uuid.New().String()
			transactionManager.TransactionChannel <- StartTransaction{TransactionID: txID}
			assertTransaction(t, transactionManager, txID, InProgress, map[string]*Action{})

			actions := tc.operation(txID, t, transactionManager)

			assertTransaction(t, transactionManager, txID, Failed, actions)
		})
	}

	defer transactionManager.Stop()

	// sleep and wait for automatic cleanup to remove the successful transaction
	time.Sleep(1 * time.Second)
	assert.Len(t, transactionManager.Transactions, 0)
}

func TestTransactionManager_TransactionTimeout(t *testing.T) {
	staleTimeout := 50 * time.Millisecond
	transactionManager := MakeTransactionManager(100, 10*time.Millisecond, staleTimeout,
		500*time.Millisecond)

	transactionManager.Start()

	for _, tc := range []struct {
		testCase  string
		operation func(txID string, t *testing.T, manager *TransactionManager) map[string]*Action
	}{
		{
			testCase: "Transaction timeout after becoming stale (no actions)",
			operation: func(txID string, t *testing.T, manager *TransactionManager) (actions map[string]*Action) {
				return
			},
		},
		{
			testCase: "Transaction timeout after becoming stale for a second time with actions",
			operation: func(txID string, t *testing.T, manager *TransactionManager) map[string]*Action {
				// assert that we have a InProgress transaction
				actions := make(map[string]*Action, 0)
				assertTransaction(t, transactionManager, txID, InProgress, actions)

				// sleep for staleTimeout and assert that the transaction has become stale
				time.Sleep(staleTimeout)
				assertTransaction(t, transactionManager, txID, Stale, actions)

				// commit an action and assert that the transaction is again in progress
				actID := uuid.New().String()
				commitAssertAction(t, transactionManager, txID, actID, actions)

				assertTransaction(t, transactionManager, txID, InProgress, actions)

				return actions
			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			txID := uuid.New().String()
			transactionManager.TransactionChannel <- StartTransaction{TransactionID: txID}
			assertTransaction(t, transactionManager, txID, InProgress, map[string]*Action{})

			actions := tc.operation(txID, t, transactionManager)

			time.Sleep(staleTimeout)

			assertTransaction(t, transactionManager, txID, Stale, actions)
		})
	}

	defer transactionManager.Stop()

	// sleep and wait for automatic cleanup to remove the successful transaction
	time.Sleep(1 * time.Second)
	assert.Len(t, transactionManager.Transactions, 0)
}

func TestTransactionManager_ErrorHandling(t *testing.T) {
	transactionManager := MakeTransactionManager(100, 100*time.Millisecond, 500*time.Millisecond,
		500*time.Millisecond)

	for _, tc := range []struct {
		testCase  string
		operation func(t *testing.T, manager *TransactionManager)
	}{
		{
			testCase: "Transaction created before starting transaction manager",
			operation: func(t *testing.T, manager *TransactionManager) {
				txID := uuid.New().String()
				transactionManager.TransactionChannel <- StartTransaction{TransactionID: txID}

				// assert that the transaction manager is not running and that we have no transactions (nothing broke)
				assert.False(t, transactionManager.running)
				assert.Empty(t, transactionManager.Transactions)

				// start the transaction manager for the following tests
				transactionManager.Start()

				assert.True(t, transactionManager.running)
				assert.Len(t, transactionManager.Transactions, 0)
			},
		},
		{
			testCase: "Commit action for a non-existing transaction (FLAKY)",
			operation: func(t *testing.T, manager *TransactionManager) {
				txID := uuid.New().String()
				actID := uuid.New().String()
				transactionManager.TransactionChannel <- CommitAction{TransactionID: txID, ActionID: actID}

				// assert nothing was created and that the transaction manager is still running
				assert.True(t, transactionManager.running)
				assert.Len(t, transactionManager.Transactions, 0)
			},
		},
		{
			testCase: "Acknowledge a non-existing action for a transaction",
			operation: func(t *testing.T, manager *TransactionManager) {
				txID := uuid.New().String()
				actions := make(map[string]*Action, 0)

				transactionManager.TransactionChannel <- StartTransaction{TransactionID: txID}
				assertTransaction(t, transactionManager, txID, InProgress, actions)

				actID := uuid.New().String()
				commitAssertAction(t, transactionManager, txID, actID, actions)

				transactionManager.TransactionChannel <- AckAction{TransactionID: txID, ActionID: "non-existing-action"}
				assertTransaction(t, transactionManager, txID, InProgress, actions)

			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			tc.operation(t, transactionManager)
		})
	}

	defer transactionManager.Stop()

	// sleep and wait for automatic cleanup to remove the successful transaction
	time.Sleep(1 * time.Second)
	assert.Len(t, transactionManager.Transactions, 0)
}

func assertTransaction(t *testing.T, transactionManager *TransactionManager, txID string, state TransactionState,
	actions map[string]*Action) {
	// give the transaction manager a bit of time to insert the transaction before running the assertion
	time.Sleep(10 * time.Millisecond)
	transaction, found := transactionManager.Transactions[txID]
	assert.True(t, found, "Transaction %s not found in the transaction map", txID)
	assert.Equal(t, txID, transaction.TransactionID)
	assert.Equal(t, state, transaction.State)
	assert.Equal(t, len(actions), len(transaction.Actions))
	for _, action := range transaction.Actions {
		expectedAction, found := actions[action.ActionID]
		assert.True(t, found)
		assert.Equal(t, expectedAction.ActionID, action.ActionID)
		assert.Equal(t, expectedAction.Acknowledged, action.Acknowledged)
	}
}

func commitAssertAction(t *testing.T, transactionManager *TransactionManager, txID, actID string, actions map[string]*Action) {
	transactionManager.TransactionChannel <- CommitAction{TransactionID: txID, ActionID: actID}
	actions[actID] = &Action{ActionID: actID, Acknowledged: false}
	assertTransaction(t, transactionManager, txID, InProgress, actions)
}
