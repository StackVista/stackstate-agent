package handler

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/state"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionforwarder"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/config"
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

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	config.Datadog.Set("sts_url", httpServer.URL)
	config.Datadog.Set("api_key", "my-test-api-key")
	config.Datadog.Set("transactional_forwarder_retry_min", 100*time.Millisecond)
	config.Datadog.Set("transactional_forwarder_retry_max", 500*time.Millisecond)

	state.InitCheckStateManager()
	transactionforwarder.InitTransactionalForwarder()
	transactionbatcher.InitTransactionalBatcher("test-hostname", "test-agent-name", 100,
		10*time.Second)
	transactionmanager.InitTransactionManager(100, 250*time.Millisecond, 500*time.Millisecond,
		500*time.Millisecond)

	ch := NewCheckHandler(&check.STSTestCheck{Name: "my-check-handler-transactional-state-check"}, &check.TestCheckReloader{},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0}).(*checkHandler)

	stateKey := fmt.Sprintf("%s:state", ch.String())
	expectedState := "{\"state\":\"transactional\"}"

	ch.StartTransaction()

	ch.SetStateTransactional(stateKey, expectedState)

	assert.Equal(t, "{}", ch.GetState(stateKey))

	ch.StopTransaction()

	time.Sleep(250 * time.Millisecond)

	assert.Equal(t, expectedState, ch.GetState(stateKey))

	err := os.RemoveAll("./testdata/my-check-handler-transactional-state-check")
	assert.NoError(t, err, "error cleaning up test data file")

	// stop the transactional components
	transactionforwarder.GetTransactionalForwarder().Stop()
	transactionbatcher.GetTransactionalBatcher().Stop()
	transactionmanager.GetTransactionManager().Stop()
}
