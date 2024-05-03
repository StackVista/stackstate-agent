package transactionforwarder

import (
	"github.com/DataDog/datadog-agent/pkg/collector/transactional"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional/transactionmanager"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

var (
	testTransactionID  = "transaction1"
	testTransaction2ID = "transaction2"
	testActionID       = "action1"
	testActionID2      = "action2"
)

func TestForwarder(t *testing.T) {
	var attemptCounter int32
	maxRetries := int32(5)

	httpServer := func(retries int32) *httptest.Server {
		return httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				atomic.AddInt32(&attemptCounter, 1)

				currentAttempts := atomic.LoadInt32(&attemptCounter)
				if currentAttempts >= retries {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}),
		)
	}

	rejectedActionAssertion := func(manager *transactionmanager.MockTransactionManager,
		txMap map[string]transactional.PayloadTransaction) {
		for i := 0; i < len(txMap); i++ {
			rejectAction := manager.NextAction().(transactionmanager.RejectAction)
			expectedPT, found := txMap[rejectAction.TransactionID]
			if !found {
				assert.Fail(t, "Commit action for transaction %s, not found in expected transaction state: %v",
					rejectAction.TransactionID, txMap)
			}
			assert.Equal(t, expectedPT.ActionID, rejectAction.ActionID)
			assert.Contains(t, rejectAction.Reason, "/intake?api_key=my-test-api-key giving up after 5 attempt(s)")
		}
	}

	for _, tc := range []struct {
		TestCase                     string
		TestTransactionalPayload     TransactionalPayload
		TransactionManagerAssertions func(manager *transactionmanager.MockTransactionManager,
			txMap map[string]transactional.PayloadTransaction)
		Attempts int32
	}{
		{
			TestCase: "Failed HTTP request after retries, reject single action",
			TestTransactionalPayload: TransactionalPayload{
				Path: transactional.IntakePath,
				TransactionActionMap: map[string]transactional.PayloadTransaction{
					testTransactionID: {
						ActionID:             testActionID,
						CompletedTransaction: false,
					},
				},
			},
			TransactionManagerAssertions: rejectedActionAssertion,
			Attempts:                     maxRetries + 1,
		},
		{
			TestCase: "Failed HTTP request after retries, reject multiple actions",
			TestTransactionalPayload: TransactionalPayload{
				Path: transactional.IntakePath,
				TransactionActionMap: map[string]transactional.PayloadTransaction{
					testTransactionID: {
						ActionID:             testActionID,
						CompletedTransaction: false,
					},
					testTransaction2ID: {
						ActionID:             testActionID2,
						CompletedTransaction: false,
					},
				},
			},
			TransactionManagerAssertions: rejectedActionAssertion,
			Attempts:                     maxRetries + 1,
		},
		{
			TestCase: "Successful HTTP request after 3 attempts, expect AcknowledgeAction for each in " +
				"progress transaction.",
			TestTransactionalPayload: TransactionalPayload{
				Path: transactional.IntakePath,
				TransactionActionMap: map[string]transactional.PayloadTransaction{
					testTransactionID: {
						ActionID:             testActionID,
						CompletedTransaction: false,
					},
					testTransaction2ID: {
						ActionID:             testActionID2,
						CompletedTransaction: false,
					},
				},
			},
			TransactionManagerAssertions: func(manager *transactionmanager.MockTransactionManager,
				txMap map[string]transactional.PayloadTransaction) {
				for i := 0; i < len(txMap); i++ {
					ackAction := manager.NextAction().(transactionmanager.AckAction)
					expectedPT, found := txMap[ackAction.TransactionID]
					if !found {
						assert.Fail(t, "Commit action for transaction %s, not found in expected transaction state: %v",
							ackAction.TransactionID, txMap)
					}
					assert.Equal(t, expectedPT.ActionID, ackAction.ActionID)
				}
			},
			Attempts: 3,
		},
		{
			TestCase: "Successful HTTP request after 1 attempt, expect AcknowledgeAction + " +
				"CompleteTransaction for each completed transaction.",
			TestTransactionalPayload: TransactionalPayload{
				Path: transactional.IntakePath,
				TransactionActionMap: map[string]transactional.PayloadTransaction{
					testTransactionID: {
						ActionID:             testActionID,
						CompletedTransaction: true,
					},
					testTransaction2ID: {
						ActionID:             testActionID2,
						CompletedTransaction: true,
					},
				},
			},
			TransactionManagerAssertions: func(manager *transactionmanager.MockTransactionManager,
				txMap map[string]transactional.PayloadTransaction) {
				for i := 0; i < len(txMap); i++ {
					ackAction := manager.NextAction().(transactionmanager.AckAction)
					expectedPT, found := txMap[ackAction.TransactionID]
					if !found {
						assert.Fail(t, "Commit action for transaction %s, not found in expected transaction state: %v",
							ackAction.TransactionID, txMap)
					}
					assert.Equal(t, expectedPT.ActionID, ackAction.ActionID)

					completedTx := manager.NextAction().(transactionmanager.CompleteTransaction)
					assert.Equal(t, ackAction.TransactionID, completedTx.TransactionID)
				}
			},
			Attempts: 1,
		},
		{
			TestCase: "Expect AcknowledgeAction + CompleteTransaction for each completed transaction while sending " +
				"an empty HTTP request.",
			TestTransactionalPayload: TransactionalPayload{
				Path: transactional.IntakePath,
				TransactionActionMap: map[string]transactional.PayloadTransaction{
					testTransactionID: {
						ActionID:             testActionID,
						CompletedTransaction: true,
					},
				},
			},
			TransactionManagerAssertions: func(manager *transactionmanager.MockTransactionManager,
				txMap map[string]transactional.PayloadTransaction) {
				for i := 0; i < len(txMap); i++ {
					ackAction := manager.NextAction().(transactionmanager.AckAction)
					expectedPT, found := txMap[ackAction.TransactionID]
					if !found {
						assert.Fail(t, "Commit action for transaction %s, not found in expected transaction state: %v",
							ackAction.TransactionID, txMap)
					}
					assert.Equal(t, expectedPT.ActionID, ackAction.ActionID)

					completedTx := manager.NextAction().(transactionmanager.CompleteTransaction)
					assert.Equal(t, ackAction.TransactionID, completedTx.TransactionID)
				}
			},
			Attempts: 1,
		},
	} {
		t.Run(tc.TestCase, func(t *testing.T) {
			// initial attempt counter to 0
			atomic.StoreInt32(&attemptCounter, 0)

			tm := transactionmanager.NewMockTransactionManager()

			server := httpServer(tc.Attempts)

			config.Datadog.SetWithoutSource("sts_url", server.URL)
			config.Datadog.SetWithoutSource("api_key", "my-test-api-key")
			config.Datadog.SetWithoutSource("transactional_forwarder_retry_min", 100*time.Millisecond)
			config.Datadog.SetWithoutSource("transactional_forwarder_retry_max", 500*time.Millisecond)

			fwd := newTransactionalForwarder()

			fwd.SubmitTransactionalIntake(tc.TestTransactionalPayload)

			tc.TransactionManagerAssertions(tm, tc.TestTransactionalPayload.TransactionActionMap)

			if tc.Attempts > maxRetries {
				assert.Equal(t, maxRetries, atomic.LoadInt32(&attemptCounter))
			} else {
				assert.Equal(t, tc.Attempts, atomic.LoadInt32(&attemptCounter))
			}

			// reset attempt counter to 0
			server.Close()
			atomic.StoreInt32(&attemptCounter, 0)
			fwd.Stop()

		})
	}
}

func TestForwarder_Multiple(t *testing.T) {
	httpServer := func() *httptest.Server {
		return httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		)
	}

	manager := transactionmanager.NewMockTransactionManager()

	server := httpServer()

	config.Datadog.SetWithoutSource("sts_url", server.URL)
	config.Datadog.SetWithoutSource("api_key", "my-test-api-key")
	config.Datadog.SetWithoutSource("transactional_forwarder_retry_min", 100*time.Millisecond)
	config.Datadog.SetWithoutSource("transactional_forwarder_retry_max", 500*time.Millisecond)

	fwd := newTransactionalForwarder()

	for i := 0; i < 5; i++ {
		txMap := map[string]transactional.PayloadTransaction{
			testTransactionID: {
				ActionID:             testActionID,
				CompletedTransaction: true,
			},
		}

		testTransactionPayload := TransactionalPayload{
			Path:                 transactional.IntakePath,
			TransactionActionMap: txMap,
		}

		fwd.SubmitTransactionalIntake(testTransactionPayload)

		ackAction := manager.NextAction().(transactionmanager.AckAction)
		expectedPT, found := txMap[ackAction.TransactionID]
		if !found {
			assert.Fail(t, "Commit action for transaction %s, not found in expected transaction state: %v",
				ackAction.TransactionID, txMap)
		}
		assert.Equal(t, expectedPT.ActionID, ackAction.ActionID)

		completedTx := manager.NextAction().(transactionmanager.CompleteTransaction)
		assert.Equal(t, ackAction.TransactionID, completedTx.TransactionID)
	}

	for i := 0; i < 5; i++ {
		txMap := map[string]transactional.PayloadTransaction{
			testTransactionID: {
				ActionID:             testActionID,
				CompletedTransaction: true,
			},
		}

		testTransactionPayload := TransactionalPayload{
			Path:                 transactional.IntakePath,
			TransactionActionMap: txMap,
		}

		fwd.SubmitTransactionalIntake(testTransactionPayload)

		ackAction := manager.NextAction().(transactionmanager.AckAction)
		expectedPT, found := txMap[ackAction.TransactionID]
		if !found {
			assert.Fail(t, "Commit action for transaction %s, not found in expected transaction state: %v",
				ackAction.TransactionID, txMap)
		}
		assert.Equal(t, expectedPT.ActionID, ackAction.ActionID)

		completedTx := manager.NextAction().(transactionmanager.CompleteTransaction)
		assert.Equal(t, ackAction.TransactionID, completedTx.TransactionID)
	}

	for i := 0; i < 5; i++ {
		txMap := map[string]transactional.PayloadTransaction{
			testTransactionID: {
				ActionID:             testActionID,
				CompletedTransaction: true,
			},
		}

		testTransactionPayload := TransactionalPayload{
			Path:                 transactional.IntakePath,
			TransactionActionMap: txMap,
		}

		fwd.SubmitTransactionalIntake(testTransactionPayload)

		ackAction := manager.NextAction().(transactionmanager.AckAction)
		expectedPT, found := txMap[ackAction.TransactionID]
		if !found {
			assert.Fail(t, "Commit action for transaction %s, not found in expected transaction state: %v",
				ackAction.TransactionID, txMap)
		}
		assert.Equal(t, expectedPT.ActionID, ackAction.ActionID)

		completedTx := manager.NextAction().(transactionmanager.CompleteTransaction)
		assert.Equal(t, ackAction.TransactionID, completedTx.TransactionID)
	}

}
