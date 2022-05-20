package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	instance = topology.Instance{
		Type: "mytype",
		URL:  "myurl",
	}

	testComponent = topology.Component{
		ExternalID: "id",
		Type:       topology.Type{Name: "typename"},
		Data:       map[string]interface{}{},
	}
)

func TestCheckHandlerAPI(t *testing.T) {
	// init global transactionbatcher used by the check no handler
	mockBatcher := transactionbatcher.NewMockTransactionalBatcher()
	transactionmanager.NewMockTransactionManager()

	checkHandler := NewCheckHandler(&check.STSTestCheck{Name: "my-check-handler-test-check"}, &check.TestCheckReloader{},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0})

	transactionID := checkHandler.SubmitStartTransaction()

	time.Sleep(100 * time.Millisecond)
	checkHandler.SubmitComponent(instance, testComponent)

	state := mockBatcher.CollectedTopology.Flush()

	expectedState := transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"my-check-handler-test-check": {
				Transaction: &transactionbatcher.BatchTransaction{
					TransactionID:        transactionID,
					CompletedTransaction: false,
				},
				Topology: &topology.Topology{
					StartSnapshot: false,
					StopSnapshot:  false,
					Instance:      instance,
					Components: []topology.Component{
						testComponent,
					},
					Relations: []topology.Relation{},
					DeleteIDs: []string{},
				},
				Health: map[string]health.Health{},
			},
		})

	assert.EqualValues(t, expectedState, state)

}
