package handler

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/state"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestCheckHandlerNonTransactionalAPI(t *testing.T) {
	testCheck := &check.STSTestCheck{Name: "my-check-handler-non-transactional-check"}
	nonTransactionCH := MakeNonTransactionalCheckHandler(testCheck, &check.TestCheckReloader{},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0})

	mockBatcher := batcher.NewMockBatcher()

	nonTransactionCH.SubmitStartSnapshot(instance)
	nonTransactionCH.SubmitComponent(instance, testComponent)
	nonTransactionCH.SubmitRelation(instance, testRelation)
	nonTransactionCH.SubmitDelete(instance, deleteID)

	nonTransactionCH.SubmitHealthStartSnapshot(testStream, 1, 0)
	nonTransactionCH.SubmitHealthCheckData(testStream, testCheckData)
	nonTransactionCH.SubmitHealthStopSnapshot(testStream)

	nonTransactionCH.SubmitRawMetricsData(testRawMetricsData)
	nonTransactionCH.SubmitRawMetricsData(testRawMetricsData2)

	nonTransactionCH.SubmitStopSnapshot(instance)

	actualState := mockBatcher.CollectedTopology.Flush()

	expectedState := batcher.CheckInstanceBatchStates(map[check.ID]batcher.CheckInstanceBatchState{
		nonTransactionCH.ID(): {
			Topology: &topology.Topology{
				StartSnapshot: true,
				StopSnapshot:  true,
				Instance:      instance,
				Components:    []topology.Component{testComponent},
				Relations:     []topology.Relation{testRelation},
				DeleteIDs:     []string{deleteID},
			},
			Metrics: &[]telemetry.RawMetrics{
				testRawMetricsData,
				testRawMetricsData2,
			},
			Health: map[string]health.Health{
				testStream.GoString(): {
					StartSnapshot: testStartSnapshot,
					StopSnapshot:  testStopSnapshot,
					Stream:        testStream,
					CheckStates:   []health.CheckData{testCheckData},
				},
			},
		},
	})

	assert.Equal(t, expectedState, actualState)

	mockBatcher.Shutdown()

}

func TestNonTransactionalCheckHandler_StartTransaction(t *testing.T) {
	InitCheckManager(CheckNoReloader{})
	transactionbatcher.NewMockTransactionalBatcher()
	transactionmanager.NewMockTransactionManager()

	testCheck := &check.STSTestCheck{Name: "my-check-handler-non-transactional-check"}
	GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{1, 2, 3}, integration.Data{0, 0, 0})

	checkHandler := GetCheckManager().GetCheckHandler(testCheck.ID())

	assert.Equal(t, "NonTransactionalCheckHandler", checkHandler.Name())

	transactionID := checkHandler.StartTransaction()

	time.Sleep(100 * time.Millisecond) // sleep to give everything a bit of time to finish up

	transactionalCheckHandler := GetCheckManager().GetCheckHandler(testCheck.ID()).(*TransactionalCheckHandler)
	assert.Equal(t, "TransactionalCheckHandler", transactionalCheckHandler.Name())
	assert.Equal(t, transactionID, transactionalCheckHandler.GetCurrentTransaction())

	// stop the transactional components
	transactionbatcher.GetTransactionalBatcher().Stop()
	transactionmanager.GetTransactionManager().Stop()

	GetCheckManager().Stop()

}

func TestNonTransactionalCheckHandler_State(t *testing.T) {
	os.Setenv("DD_CHECK_STATE_ROOT_PATH", "./testdata")
	state.InitCheckStateManager()

	testCheck := &check.STSTestCheck{Name: "my-check-handler-non-transactional-check"}
	nonTransactionCH := MakeNonTransactionalCheckHandler(testCheck, &check.TestCheckReloader{},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0})

	stateKey := fmt.Sprintf("%s:state", testCheck.Name)

	actualState := nonTransactionCH.GetState(stateKey)
	expectedState := "{\"non\":\"transactional\"}"
	assert.Equal(t, expectedState, actualState)

	checkState, err := state.GetCheckStateManager().GetState(stateKey)
	assert.NoError(t, err, "unexpected error occurred when trying to get state for", stateKey)
	assert.Equal(t, expectedState, checkState)

	updatedState := "{\"e\":\"f\"}"
	nonTransactionCH.SetState(stateKey, updatedState)

	checkState, err = state.GetCheckStateManager().GetState(stateKey)
	assert.NoError(t, err, "unexpected error occurred when trying to get state for", stateKey)
	assert.Equal(t, updatedState, checkState)

	// reset to original
	nonTransactionCH.SetState(stateKey, expectedState)
}
