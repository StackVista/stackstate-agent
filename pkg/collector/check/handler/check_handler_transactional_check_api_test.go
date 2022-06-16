package handler

import (
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
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

	testRelation = topology.Relation{
		ExternalID: "id2",
		Type:       topology.Type{Name: "typename"},
		SourceID:   "source",
		TargetID:   "target",
		Data:       map[string]interface{}{},
	}

	deleteID          = "my-delete-id"
	testStream        = health.Stream{Urn: "urn", SubStream: "bla"}
	testStartSnapshot = &health.StartSnapshotMetadata{ExpiryIntervalS: 0, RepeatIntervalS: 1}
	testStopSnapshot  = &health.StopSnapshotMetadata{}
	testCheckData     = health.CheckData{Unstructured: map[string]interface{}{}}

	testRawMetricsData = telemetry.RawMetrics{
		Name:      "name",
		Timestamp: 1400000,
		HostName:  "Hostname",
		Value:     200,
		Tags: []string{
			"foo",
			"bar",
		},
	}

	testRawMetricsData2 = telemetry.RawMetrics{
		Name:      "name",
		Timestamp: 1500000,
		HostName:  "hostname",
		Value:     100,
		Tags: []string{
			"hello",
			"world",
		},
	}
)

// Each table test mutates the shared checkInstanceBatchState, so running individual table tests will not produce the
// expected result. This should be run as a single test with a sequence of steps.
func TestCheckHandlerAPI(t *testing.T) {
	// init global transactionbatcher used by the check no handler
	mockBatcher := transactionbatcher.NewMockTransactionalBatcher()
	transactionmanager.NewMockTransactionManager()

	checkHandler := NewTransactionalCheckHandler(&check.STSTestCheck{Name: "my-check-handler-test-check"}, &check.TestCheckReloader{},
		integration.Data{1, 2, 3}, integration.Data{0, 0, 0})
	var transactionID string
	checkInstanceBatchState := &transactionbatcher.TransactionCheckInstanceBatchState{}

	for _, tc := range []struct {
		testCase             string
		checkHandlerFunction func(handler CheckHandler)
		stateMutation        func(*transactionbatcher.TransactionCheckInstanceBatchState)
	}{
		{
			testCase: "Start transaction should produce a batch transaction in the TransactionCheckInstanceBatchState",
			checkHandlerFunction: func(handler CheckHandler) {
				transactionID = handler.StartTransaction()
			},
			stateMutation: func(state *transactionbatcher.TransactionCheckInstanceBatchState) {
				state.Transaction = &transactionbatcher.BatchTransaction{
					TransactionID:        transactionID,
					CompletedTransaction: false,
				}
				state.Health = make(map[string]health.Health, 0)
			},
		},
		{
			testCase: "Submit start snapshot should produce a topology snapshot in the TransactionCheckInstanceBatchState",
			checkHandlerFunction: func(handler CheckHandler) {
				handler.SubmitStartSnapshot(instance)
			},
			stateMutation: func(state *transactionbatcher.TransactionCheckInstanceBatchState) {
				state.Topology = &topology.Topology{
					StartSnapshot: true,
					StopSnapshot:  false,
					Instance:      instance,
					Components:    make([]topology.Component, 0),
					Relations:     make([]topology.Relation, 0),
					DeleteIDs:     make([]string, 0),
				}
			},
		},
		{
			testCase: "Submit component should produce a component in the TransactionCheckInstanceBatchState",
			checkHandlerFunction: func(handler CheckHandler) {
				handler.SubmitComponent(instance, testComponent)
			},
			stateMutation: func(state *transactionbatcher.TransactionCheckInstanceBatchState) {
				state.Topology.Components = append(state.Topology.Components, testComponent)
			},
		},
		{
			testCase: "Submit relation should produce a relation in the TransactionCheckInstanceBatchState",
			checkHandlerFunction: func(handler CheckHandler) {
				handler.SubmitRelation(instance, testRelation)
			},
			stateMutation: func(state *transactionbatcher.TransactionCheckInstanceBatchState) {
				state.Topology.Relations = append(state.Topology.Relations, testRelation)
			},
		},
		{
			testCase: "Submit delete should produce a delete id in the TransactionCheckInstanceBatchState",
			checkHandlerFunction: func(handler CheckHandler) {
				handler.SubmitDelete(instance, deleteID)
			},
			stateMutation: func(state *transactionbatcher.TransactionCheckInstanceBatchState) {
				state.Topology.DeleteIDs = append(state.Topology.DeleteIDs, deleteID)
			},
		},
		{
			testCase: "Submit stop snapshot should produce a topology snapshot in the TransactionCheckInstanceBatchState",
			checkHandlerFunction: func(handler CheckHandler) {
				handler.SubmitStopSnapshot(instance)
			},
			stateMutation: func(state *transactionbatcher.TransactionCheckInstanceBatchState) {
				state.Topology.StopSnapshot = true
			},
		},
		{
			testCase: "Submit start health snapshot should produce a health snapshot in the TransactionCheckInstanceBatchState",
			checkHandlerFunction: func(handler CheckHandler) {
				handler.SubmitHealthStartSnapshot(testStream, 1, 0)
			},
			stateMutation: func(state *transactionbatcher.TransactionCheckInstanceBatchState) {
				state.Health[testStream.GoString()] = health.Health{
					StartSnapshot: testStartSnapshot,
					StopSnapshot:  nil,
					Stream:        testStream,
					CheckStates:   make([]health.CheckData, 0),
				}
			},
		},
		{
			testCase: "Submit health check data should append to the health check states in the TransactionCheckInstanceBatchState",
			checkHandlerFunction: func(handler CheckHandler) {
				handler.SubmitHealthCheckData(testStream, testCheckData)
			},
			stateMutation: func(state *transactionbatcher.TransactionCheckInstanceBatchState) {
				h := state.Health[testStream.GoString()]
				h.CheckStates = append(state.Health[testStream.GoString()].CheckStates, testCheckData)
				state.Health[testStream.GoString()] = h
			},
		},
		{
			testCase: "Submit stop health snapshot should produce a health snapshot in the TransactionCheckInstanceBatchState",
			checkHandlerFunction: func(handler CheckHandler) {
				handler.SubmitHealthStopSnapshot(testStream)
			},
			stateMutation: func(state *transactionbatcher.TransactionCheckInstanceBatchState) {
				h := state.Health[testStream.GoString()]
				h.StopSnapshot = testStopSnapshot
				state.Health[testStream.GoString()] = h
			},
		},
		{
			testCase: "Submit raw metric data should produce a raw metric in the TransactionCheckInstanceBatchState",
			checkHandlerFunction: func(handler CheckHandler) {
				handler.SubmitRawMetricsData(testRawMetricsData)
			},
			stateMutation: func(state *transactionbatcher.TransactionCheckInstanceBatchState) {
				state.Metrics = &telemetry.Metrics{Values: []telemetry.RawMetrics{testRawMetricsData}}
			},
		},
		{
			testCase: "Stop transaction should mark a batch transaction as complete in the TransactionCheckInstanceBatchState",
			checkHandlerFunction: func(handler CheckHandler) {
				handler.StopTransaction()
			},
			stateMutation: func(state *transactionbatcher.TransactionCheckInstanceBatchState) {
				state.Transaction.CompletedTransaction = true
			},
		},
	} {
		t.Run(tc.testCase, func(t *testing.T) {
			tc.checkHandlerFunction(checkHandler)
			tc.stateMutation(checkInstanceBatchState)

			time.Sleep(100 * time.Millisecond)

			actualState, found := mockBatcher.GetCheckState(checkHandler.ID())
			assert.True(t, found, "check state for %s was not found", checkHandler.ID())
			assert.EqualValues(t, *checkInstanceBatchState, actualState)
		})
	}

	// test check handler cancel transaction
	t.Run("check handler cancel transaction", func(t *testing.T) {
		ch := NewTransactionalCheckHandler(&check.STSTestCheck{Name: "my-check-handler-cancel-transaction"},
			&check.TestCheckReloader{}, integration.Data{1, 2, 3}, integration.Data{0, 0, 0})

		txID := ch.StartTransaction()
		ch.SubmitComponent(instance, testComponent)

		batchState := transactionbatcher.TransactionCheckInstanceBatchState{
			Transaction: &transactionbatcher.BatchTransaction{
				TransactionID:        txID,
				CompletedTransaction: false,
			},
			Topology: &topology.Topology{
				StartSnapshot: false,
				StopSnapshot:  false,
				Instance:      instance,
				Components:    []topology.Component{testComponent},
				Relations:     []topology.Relation{},
				DeleteIDs:     []string{},
			},
			Health: map[string]health.Health{},
		}

		time.Sleep(100 * time.Millisecond)

		actualState, found := mockBatcher.GetCheckState(ch.ID())
		assert.True(t, found, "check state for %s was not found", ch.ID())
		assert.EqualValues(t, batchState, actualState)

		ch.CancelTransaction("test cancel transaction")

		time.Sleep(100 * time.Millisecond)
		postCancelState, found := mockBatcher.GetCheckState(ch.ID())
		assert.False(t, found, "check state for %s that should be cleaned up was found: %v", ch.ID(),
			postCancelState.JSONString())
	})

	// test check handler submit complete
	t.Run("check handler submit complete", func(t *testing.T) {
		ch := NewTransactionalCheckHandler(&check.STSTestCheck{Name: "my-check-handler-submit-complete"},
			&check.TestCheckReloader{}, integration.Data{1, 2, 3}, integration.Data{0, 0, 0})

		txID := ch.StartTransaction()
		ch.SubmitComponent(instance, testComponent)
		ch.SubmitComplete()

		batchState := transactionbatcher.TransactionCheckInstanceBatchState{
			Transaction: &transactionbatcher.BatchTransaction{
				TransactionID:        txID,
				CompletedTransaction: false,
			},
			Topology: &topology.Topology{
				StartSnapshot: false,
				StopSnapshot:  false,
				Instance:      instance,
				Components:    []topology.Component{testComponent},
				Relations:     []topology.Relation{},
				DeleteIDs:     []string{},
			},
			Health: map[string]health.Health{},
		}

		time.Sleep(100 * time.Millisecond)

		actualState, found := mockBatcher.GetCheckState(ch.ID())
		assert.True(t, found, "check state for %s was not found", ch.ID())
		assert.EqualValues(t, batchState, actualState)
	})

}
