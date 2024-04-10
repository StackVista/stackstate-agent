package handler

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	"github.com/DataDog/datadog-agent/pkg/collector/check/state"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional/transactionforwarder"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional/transactionmanager"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/stretchr/testify/assert"
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

	os.Setenv("DD_CHECK_STATE_ROOT_PATH", "./testdata")

	config.Datadog.Set("api_key", "my-test-api-key")
	config.Datadog.Set("transactional_forwarder_retry_min", 100*time.Millisecond)
	config.Datadog.Set("transactional_forwarder_retry_max", 500*time.Millisecond)

	state.InitCheckStateManager()
	transactionbatcher.InitTransactionalBatcher("test-hostname", "test-agent-name", 100)
	transactionmanager.InitTransactionManager(100, 250*time.Millisecond, 500*time.Millisecond,
		500*time.Millisecond)

	t.Run("set transaction state successful scenario", func(t *testing.T) {
		httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		config.Datadog.Set("sts_url", httpServer.URL)

		transactionforwarder.InitTransactionalForwarder()

		ch := NewTransactionalCheckHandler(&check.STSTestCheck{Name: "my-check-handler-transactional-state-check"},
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

		// stop the transactional components
		transactionforwarder.GetTransactionalForwarder().Stop()
	})

	t.Run("set transaction state success -> failure -> success scenario", func(t *testing.T) {
		httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		config.Datadog.Set("sts_url", httpServer.URL)

		transactionforwarder.InitTransactionalForwarder()

		ch := NewTransactionalCheckHandler(&check.STSTestCheck{Name: "my-check-handler-transactional-state-failure-check"},
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

		// close the http server and try to update the transactional state
		httpServer.Close()
		runScenario(updatedState, baseState, baseState)

		// assert that the state has not been updated, ie the base state from the 1st successful run is still the
		// current state
		runScenario("{\"state\":\"not-updated\"}", baseState, baseState)

		// restart the transactional forwarder, set up a new http server to make successful requests
		transactionforwarder.GetTransactionalForwarder().Stop()
		httpServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		config.Datadog.Set("sts_url", httpServer.URL)
		transactionforwarder.InitTransactionalForwarder()

		// update the base state to updatedState with the new http server
		runScenario(updatedState, baseState, updatedState)

		// assert that the state was updated to updatedState, update it to wasUpdatedState
		runScenario(wasUpdatedState, updatedState, wasUpdatedState)

		err = os.RemoveAll(fmt.Sprintf("%s/%s", config.Datadog.Get("check_state_root_path"), ch.ID()))
		assert.NoError(t, err, "error cleaning up test data file")

		// stop the transactional components
		transactionforwarder.GetTransactionalForwarder().Stop()
	})

	// stop the transactional components
	transactionforwarder.GetTransactionalForwarder().Stop()
	transactionbatcher.GetTransactionalBatcher().Stop()
	transactionmanager.GetTransactionManager().Stop()

}
