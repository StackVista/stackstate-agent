package handler

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check/state"
	"github.com/DataDog/datadog-agent/pkg/collector/check/test"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/httpclient"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionforwarder"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionmanager"
	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestCheckHandler_State_Transactional_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping check handler transactional state integration test")
	}

	t.Run("set transaction state successful scenario", func(t *testing.T) {
		httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		os.Setenv("DD_CHECK_STATE_ROOT_PATH", "./testdata")

		stateManager := state.NewCheckStateManager()
		defer stateManager.Clear()

		manager := transactionmanager.NewTransactionManager(100, 250*time.Millisecond, 500*time.Millisecond,
			500*time.Millisecond)
		defer manager.Stop()

		client := httpclient.NewStackStateClient(
			&httpclient.ClientHost{
				APIKey:          "my-test-api-key",
				HostURL:         httpServer.URL,
				RetryWaitMin:    100 * time.Millisecond,
				RetryWaitMax:    500 * time.Millisecond,
				ContentEncoding: httpclient.IdentityContentType,
			})

		forwarder := transactionforwarder.NewTransactionalForwarder(client, manager)
		defer forwarder.Stop()

		batcher := transactionbatcher.NewTransactionalBatcher("test-hostname", 100, forwarder, manager, false)
		defer forwarder.Stop()

		ch := NewTransactionalCheckHandler(
			stateManager, batcher, manager,
			&test.STSTestCheck{Name: "my-check-handler-transactional-state-check"},
			integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*TransactionalCheckHandler)

		stateKey := fmt.Sprintf("%s:state", ch.String())
		expectedState := "{\"state\":\"transactional\"}"

		ch.StartTransaction()

		ch.SetTransactionState(stateKey, expectedState)

		assert.Equal(t, "{}", ch.GetState(stateKey))

		ch.StopTransaction()

		time.Sleep(250 * time.Millisecond)

		assert.Equal(t, expectedState, ch.GetState(stateKey))

		err := os.RemoveAll(fmt.Sprintf("%s/%s", config.Datadog.Get("check_state_root_path"), ch.ID()))
		assert.NoError(t, err, "error cleaning up test data file")
	})

	t.Run("set transaction state success -> failure -> success scenario", func(t *testing.T) {
		var statusCode atomic.Int32
		statusCode.Store(200)

		httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(int(statusCode.Load()))
		}))

		os.Setenv("DD_CHECK_STATE_ROOT_PATH", "./testdata")

		stateManager := state.NewCheckStateManager()
		defer stateManager.Clear()

		manager := transactionmanager.NewTransactionManager(100, 250*time.Millisecond, 500*time.Millisecond,
			500*time.Millisecond)
		defer manager.Stop()

		client := httpclient.NewStackStateClient(
			&httpclient.ClientHost{
				APIKey:          "my-test-api-key",
				HostURL:         httpServer.URL,
				RetryWaitMin:    100 * time.Millisecond,
				RetryWaitMax:    500 * time.Millisecond,
				ContentEncoding: httpclient.IdentityContentType,
			})

		forwarder := transactionforwarder.NewTransactionalForwarder(client, manager)
		defer forwarder.Stop()

		batcher := transactionbatcher.NewTransactionalBatcher("test-hostname", 100, forwarder, manager, false)
		defer forwarder.Stop()

		ch := NewTransactionalCheckHandler(stateManager, batcher, manager, &test.STSTestCheck{Name: "my-check-handler-transactional-state-failure-check"},
			integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*TransactionalCheckHandler)

		runScenario := func(setState, currentState, stopState string) {
			stateKey := fmt.Sprintf("%s:state", ch.String())

			ch.StartTransaction()
			ch.SubmitComponent(instance, testComponent)
			ch.SetTransactionState(stateKey, setState)

			assert.Equal(t, currentState, ch.GetState(stateKey))

			ch.StopTransaction()

			assert.Eventually(t, func() bool {
				return ch.GetState(stateKey) == stopState
			}, 100*time.Millisecond, 10*time.Millisecond)

		}

		// start with a clean state
		err := os.RemoveAll(fmt.Sprintf("%s/%s", config.Datadog.Get("check_state_root_path"), ch.ID()))
		assert.NoError(t, err, "error cleaning up test data file")

		baseState := "{\"state\":\"base\"}"
		updatedState := "{\"state\":\"updated\"}"
		wasUpdatedState := "{\"state\":\"was-updated\"}"
		runScenario(baseState, "{}", baseState)

		// Generate broken response
		statusCode.Store(500)
		runScenario(updatedState, baseState, baseState)

		// assert that the state has not been updated, ie the base state from the 1st successful run is still the
		// current state
		runScenario("{\"state\":\"not-updated\"}", baseState, baseState)

		// Fix response
		statusCode.Store(200)
		// update the base state to updatedState with the new http server
		runScenario(updatedState, baseState, updatedState)

		// assert that the state was updated to updatedState, update it to wasUpdatedState
		runScenario(wasUpdatedState, updatedState, wasUpdatedState)

		err = os.RemoveAll(fmt.Sprintf("%s/%s", config.Datadog.Get("check_state_root_path"), ch.ID()))
		assert.NoError(t, err, "error cleaning up test data file")
	})

}
