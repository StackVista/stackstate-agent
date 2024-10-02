package handler

import (
	"fmt"
	"github.com/DataDog/datadog-agent/pkg/aggregator/mocksender"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/batcher"
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	"github.com/DataDog/datadog-agent/pkg/collector/check/state"
	"github.com/DataDog/datadog-agent/pkg/collector/check/test"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/health"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/telemetry"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
	"time"
)

func TestCheckHandlerNonTransactionalAPI(t *testing.T) {
	testCheck := &test.STSTestCheck{Name: "my-check-handler-non-transactional-check"}
	mockBatcher := batcher.NewMockBatcher()

	nonTransactionCH := MakeNonTransactionalCheckHandler(
		nil, mockBatcher, state.NewCheckStateManager(),
		testCheck,
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0})

	assert.Equal(t, "NonTransactionalCheckHandler", nonTransactionCH.Name())
	// call these to ensure they cause no errors
	nonTransactionCH.DiscardTransaction("none")
	nonTransactionCH.StopTransaction()
	nonTransactionCH.SetTransactionState("", "")

	// sender for non-transactional events
	sender := mocksender.NewMockSender(testCheck.ID())
	sender.On("Event", mock.AnythingOfType("event.Event"))

	nonTransactionCH.SubmitStartSnapshot(instance)
	nonTransactionCH.SubmitComponent(instance, testComponent)
	nonTransactionCH.SubmitRelation(instance, testRelation)
	nonTransactionCH.SubmitDelete(instance, deleteID)

	nonTransactionCH.SubmitHealthStartSnapshot(testStream, 1, 0)
	nonTransactionCH.SubmitHealthCheckData(testStream, testCheckData)
	nonTransactionCH.SubmitHealthStopSnapshot(testStream)

	nonTransactionCH.SubmitRawMetricsData(testRawMetricsData)
	nonTransactionCH.SubmitRawMetricsData(testRawMetricsData2)
	nonTransactionCH.SubmitEvent(testEvent)

	nonTransactionCH.SubmitStopSnapshot(instance)

	actualState := mockBatcher.CollectedTopology.Flush()

	expectedState := batcher.CheckInstanceBatchStates(map[checkid.ID]batcher.CheckInstanceBatchState{
		nonTransactionCH.ID(): {
			Topology: &topology.Topology{
				StartSnapshot: true,
				StopSnapshot:  true,
				Instance:      instance,
				Components:    []topology.Component{testComponent},
				Relations:     []topology.Relation{testRelation},
				DeleteIDs:     []string{deleteID},
			},
			Metrics: &[]telemetry.RawMetric{
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

	sender.AssertEvent(t, testEvent, 0)

	mockBatcher.Shutdown()

}

func TestNonTransactionalCheckHandler_StartTransaction(t *testing.T) {
	manager := NewCheckManager(batcher.NewMockBatcher(), transactionbatcher.NewMockTransactionalBatcher(), transactionmanager.NewMockTransactionManager())

	testCheck := &test.STSTestCheck{Name: "my-check-handler-non-transactional-check"}
	manager.RegisterCheckHandler(testCheck, integration.Data{1, 2, 3}, integration.Data{0, 0, 0})

	checkHandler := manager.GetCheckHandler(testCheck.ID())

	assert.Equal(t, "NonTransactionalCheckHandler", checkHandler.Name())

	transactionID := checkHandler.StartTransaction()

	time.Sleep(100 * time.Millisecond) // sleep to give everything a bit of time to finish up

	transactionalCheckHandler := manager.GetCheckHandler(testCheck.ID()).(*TransactionalCheckHandler)
	assert.Equal(t, "TransactionalCheckHandler", transactionalCheckHandler.Name())
	assert.Equal(t, transactionID, transactionalCheckHandler.GetCurrentTransaction())

	manager.Stop()
}

func TestNonTransactionalCheckHandler_State(t *testing.T) {
	os.Setenv("DD_CHECK_STATE_ROOT_PATH", "./testdata")

	testCheck := &test.STSTestCheck{Name: "my-check-handler-non-transactional-check"}
	mockBatcher := batcher.NewMockBatcher()
	state := state.NewCheckStateManager()
	nonTransactionCH := MakeNonTransactionalCheckHandler(
		nil, mockBatcher, state,
		testCheck,
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0})

	stateKey := fmt.Sprintf("%s:state", testCheck.Name)

	actualState := nonTransactionCH.GetState(stateKey)
	expectedState := "{\"non\":\"transactional\"}"
	assert.Equal(t, expectedState, actualState)

	checkState, err := state.GetState(stateKey)
	assert.NoError(t, err, "unexpected error occurred when trying to get state for", stateKey)
	assert.Equal(t, expectedState, checkState)

	updatedState := "{\"e\":\"f\"}"
	nonTransactionCH.SetState(stateKey, updatedState)

	checkState, err = state.GetState(stateKey)
	assert.NoError(t, err, "unexpected error occurred when trying to get state for", stateKey)
	assert.Equal(t, updatedState, checkState)

	// reset to original
	nonTransactionCH.SetState(stateKey, expectedState)
}
