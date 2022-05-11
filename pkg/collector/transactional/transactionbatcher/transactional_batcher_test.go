package transactionbatcher

import (
	"encoding/json"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionforwarder"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	testInstance       = topology.Instance{Type: "mytype", URL: "myurl"}
	testInstance2      = topology.Instance{Type: "mytype2", URL: "myurl2"}
	testHost           = "myhost"
	testAgent          = "myagent"
	testID             = check.ID("myid")
	testID2            = check.ID("myid2")
	testTransactionID  = "transaction1"
	testTransaction2ID = "transaction2"
	testComponent      = topology.Component{
		ExternalID: "id",
		Type:       topology.Type{Name: "typename"},
		Data:       map[string]interface{}{},
	}
	testComponent2 = topology.Component{
		ExternalID: "id2",
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

	testStream        = health.Stream{Urn: "urn", SubStream: "bla"}
	testStream2       = health.Stream{Urn: "urn"}
	testStartSnapshot = &health.StartSnapshotMetadata{ExpiryIntervalS: 0, RepeatIntervalS: 1}
	testStopSnapshot  = &health.StopSnapshotMetadata{}
	testCheckData     = map[string]interface{}{}

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
		HostName:  "Hostname",
		Value:     100,
		Tags: []string{
			"hello",
			"world",
		},
	}

	testRawMetricsDataIntakeMetric  = testRawMetricsData.IntakeMetricJSON()
	testRawMetricsDataIntakeMetric2 = testRawMetricsData2.IntakeMetricJSON()
)

// TODO: these might hit nil pointers in the batcher because we only init the transaction manager and forwarder in the testBatcher function
// TODO: after the batcher operations have been executed
func testBatcher(t *testing.T, transactionState []TransactionState, expectedPayload transactional.IntakePayload) {
	fwd := transactionforwarder.NewMockTransactionalForwarder()
	tm := transactionmanager.NewMockTransactionManager()

	// get the action commit made by the batcher from the transaction manager for this payload
	commitActions := make([]transactionmanager.CommitAction, len(transactionState))
	for i, tID := range transactionState {
		commitAction := tm.NextAction().(transactionmanager.CommitAction)
		assert.Equal(t, tID.TransactionID, commitAction.TransactionID)
		commitActions[i] = commitAction
	}

	// get the intake payload that was produced for this action
	payload := fwd.NextPayload()
	actualPayload := transactional.NewIntakePayload()
	json.Unmarshal(payload.Payload, &actualPayload)

	// assert the payload matches the expected payload for the data produced
	assert.ObjectsAreEqual(expectedPayload, actualPayload)

	// assert the transaction map produced by the batcher contains the correct action id and completed status
	expectedTransactionMap := make(map[string]transactional.PayloadTransaction, len(commitActions))
	for i, ca := range commitActions {
		expectedTransactionMap[ca.TransactionID] = transactional.PayloadTransaction{
			ActionID:             ca.ActionID,
			CompletedTransaction: transactionState[i].Completed,
		}
	}

	assert.Equal(t, expectedTransactionMap, payload.TransactionActionMap)

	fwd.Stop()
	tm.Stop()
}

func TestBatchFlushSnapshotOnComplete(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100, 15*time.Second)
	batcher.SubmitStopSnapshot(testID, testTransactionID, testInstance)
	batcher.SubmitComplete(testID)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Topologies = []topology.Topology{
		{
			StartSnapshot: false,
			StopSnapshot:  true,
			Instance:      testInstance,
			Components:    []topology.Component{},
			Relations:     []topology.Relation{},
			DeleteIDs:     []string{},
		},
	}

	transactionStates := []TransactionState{
		{
			TransactionID: testTransactionID,
			Completed:     true,
		},
	}
	testBatcher(t, transactionStates, expectedPayload)

	batcher.Shutdown()
}

func TestBatchFlushHealthOnComplete(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100, 15*time.Second)

	batcher.SubmitHealthStopSnapshot(testID, testTransactionID, testStream)
	batcher.SubmitComplete(testID)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Health = []health.Health{
		{
			StopSnapshot: testStopSnapshot,
			Stream:       testStream,
			CheckStates:  []health.CheckData{},
		},
	}

	transactionStates := []TransactionState{
		{
			TransactionID: testTransactionID,
			Completed:     true,
		},
	}
	testBatcher(t, transactionStates, expectedPayload)

	batcher.Shutdown()
}

func TestBatchFlushOnComplete(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100, 15*time.Second)

	batcher.SubmitComponent(testID, testTransactionID, testInstance, testComponent)
	batcher.SubmitHealthCheckData(testID, testTransactionID, testStream, testCheckData)
	batcher.SubmitRawMetricsData(testID, testTransactionID, testRawMetricsData)
	batcher.SubmitRawMetricsData(testID, testTransactionID, testRawMetricsData2)
	batcher.SubmitComplete(testID)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Topologies = []topology.Topology{
		{
			StartSnapshot: false,
			StopSnapshot:  false,
			Instance:      testInstance,
			Components:    []topology.Component{testComponent},
			Relations:     []topology.Relation{},
			DeleteIDs:     []string{},
		},
	}
	expectedPayload.Health = []health.Health{
		{
			Stream:      testStream,
			CheckStates: []health.CheckData{testCheckData},
		},
	}
	expectedPayload.Metrics = []interface{}{testRawMetricsDataIntakeMetric, testRawMetricsDataIntakeMetric2}

	transactionStates := []TransactionState{
		{
			TransactionID: testTransactionID,
			Completed:     true,
		},
	}
	testBatcher(t, transactionStates, expectedPayload)

	batcher.Shutdown()
}

func TestBatchNoDataNoComplete(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100, 15*time.Second)

	batcher.SubmitComponent(testID, testTransactionID, testInstance, testComponent)
	batcher.SubmitComplete(testID2)

	// We now send a stop to trigger a combined commit
	batcher.SubmitStopSnapshot(testID, testTransactionID, testInstance)
	batcher.SubmitComplete(testID)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Topologies = []topology.Topology{
		{
			StartSnapshot: false,
			StopSnapshot:  true,
			Instance:      testInstance,
			Components:    []topology.Component{testComponent},
			Relations:     []topology.Relation{},
			DeleteIDs:     []string{},
		},
	}

	transactionStates := []TransactionState{
		{
			TransactionID: testTransactionID,
			Completed:     true,
		},
	}
	testBatcher(t, transactionStates, expectedPayload)

	batcher.Shutdown()
}

func TestBatchMultipleTopologiesAndHealthStreams(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100, 15*time.Second)

	batcher.SubmitStartSnapshot(testID, testTransactionID, testInstance)
	batcher.SubmitComponent(testID, testTransactionID, testInstance, testComponent)
	batcher.SubmitComponent(testID2, testTransaction2ID, testInstance2, testComponent)
	batcher.SubmitComponent(testID2, testTransaction2ID, testInstance2, testComponent)
	batcher.SubmitComponent(testID2, testTransaction2ID, testInstance2, testComponent)

	batcher.SubmitHealthStartSnapshot(testID, testTransactionID, testStream, 1, 0)
	batcher.SubmitHealthCheckData(testID, testTransactionID, testStream, testCheckData)
	batcher.SubmitHealthCheckData(testID2, testTransaction2ID, testStream2, testCheckData)

	batcher.SubmitRawMetricsData(testID, testTransactionID, testRawMetricsData)
	batcher.SubmitRawMetricsData(testID2, testTransaction2ID, testRawMetricsData)
	batcher.SubmitRawMetricsData(testID, testTransactionID, testRawMetricsData2)
	batcher.SubmitRawMetricsData(testID2, testTransaction2ID, testRawMetricsData2)

	batcher.SubmitStopSnapshot(testID, testTransactionID, testInstance)
	batcher.SubmitComplete(testID)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Topologies = []topology.Topology{
		{
			StartSnapshot: true,
			StopSnapshot:  true,
			Instance:      testInstance,
			Components:    []topology.Component{testComponent},
			Relations:     []topology.Relation{},
			DeleteIDs:     []string{},
		},
		{
			StartSnapshot: false,
			StopSnapshot:  false,
			Instance:      testInstance2,
			Components:    []topology.Component{testComponent, testComponent, testComponent},
			Relations:     []topology.Relation{},
			DeleteIDs:     []string{},
		},
	}
	expectedPayload.Health = []health.Health{
		{
			StartSnapshot: testStartSnapshot,
			Stream:        testStream,
			CheckStates:   []health.CheckData{testCheckData},
		},
		{
			Stream:      testStream2,
			CheckStates: []health.CheckData{testCheckData},
		},
	}
	expectedPayload.Metrics = []interface{}{
		testRawMetricsDataIntakeMetric,
		testRawMetricsDataIntakeMetric,
		testRawMetricsDataIntakeMetric2,
		testRawMetricsDataIntakeMetric2,
	}

	transactionStates := []TransactionState{
		{
			TransactionID: testTransactionID,
			Completed:     true,
		},
		{
			TransactionID: testTransaction2ID,
			Completed:     false,
		},
	}

	testBatcher(t, transactionStates, expectedPayload)

	batcher.Shutdown()
}

type TransactionState struct {
	TransactionID string
	Completed     bool
}

//func TestBatchFlushOnMaxElements(t *testing.T) {
//	fwd := transactionforwarder.NewMockTransactionalForwarder()
//	batcher := newTransactionalBatcher(testHost, testAgent, 2, 15*time.Second)
//
//	batcher.SubmitComponent(testID, testTransactionID, testInstance, testComponent)
//	batcher.SubmitComponent(testID, testTransactionID, testInstance, testComponent2)
//
//	message, _ := fwd.NextIntakePayload()
//
//	assert.Equal(t, message,
//		map[string]interface{}{
//			"internalHostname": "myhost",
//			"topologies": []topology.Topology{
//				{
//					StartSnapshot: false,
//					StopSnapshot:  false,
//					Instance:      testInstance,
//					Components:    []topology.Component{testComponent, testComponent2},
//					Relations:     []topology.Relation{},
//					DeleteIDs:     []string{},
//				},
//			},
//			"health":  []health.Health{},
//			"metrics": []interface{}{},
//		})
//
//	batcher.Shutdown()
//}
//
//func TestBatchFlushOnMaxHealthElements(t *testing.T) {
//	fwd := transactionforwarder.NewMockTransactionalForwarder()
//	batcher := newTransactionalBatcher(testHost, testAgent, 2, 15*time.Second)
//
//	batcher.SubmitHealthCheckData(testID, testTransactionID, testStream, testCheckData)
//	batcher.SubmitHealthCheckData(testID, testTransactionID, testStream, testCheckData)
//
//	message, _ := fwd.NextIntakePayload()
//
//	assert.Equal(t, message,
//		map[string]interface{}{
//			"internalHostname": "myhost",
//			"topologies":       []topology.Topology{},
//			"health": []health.Health{
//				{
//					Stream:      testStream,
//					CheckStates: []health.CheckData{testCheckData, testCheckData},
//				},
//			},
//			"metrics": []interface{}{},
//		})
//
//	batcher.Shutdown()
//}
//
//func TestBatchFlushOnMaxRawMetricsElements(t *testing.T) {
//	fwd := transactionforwarder.NewMockTransactionalForwarder()
//	batcher := newTransactionalBatcher(testHost, testAgent, 2, 15*time.Second)
//
//	batcher.SubmitRawMetricsData(testID, testTransactionID, testRawMetricsData)
//	batcher.SubmitRawMetricsData(testID, testTransactionID, testRawMetricsData2)
//
//	message, _ := fwd.NextIntakePayload()
//
//	assert.Equal(t, message,
//		map[string]interface{}{
//			"internalHostname": "myhost",
//			"topologies":       []topology.Topology{},
//			"health":           []health.Health{},
//			"metrics": []interface{}{
//				testRawMetricsDataIntakeMetric, testRawMetricsDataIntakeMetric2},
//		})
//
//	batcher.Shutdown()
//}
//
//func TestBatchFlushOnMaxElementsEnv(t *testing.T) {
//	fwd := transactionforwarder.NewMockTransactionalForwarder()
//
//	// set transactionbatcher max capacity via ENV var
//	os.Setenv("DD_BATCHER_CAPACITY", "1")
//	batcher := newTransactionalBatcher(testHost, testAgent, config.GetMaxCapacity(), 15*time.Second)
//
//	assert.Equal(t, 1, batcher.builder.maxCapacity)
//	batcher.SubmitComponent(testID, testTransactionID, testInstance, testComponent)
//
//	message, _ := fwd.NextIntakePayload()
//	assert.Equal(t, message,
//		map[string]interface{}{
//			"internalHostname": "myhost",
//			"topologies": []topology.Topology{
//				{
//					StartSnapshot: false,
//					StopSnapshot:  false,
//					Instance:      testInstance,
//					Components:    []topology.Component{testComponent},
//					Relations:     []topology.Relation{},
//					DeleteIDs:     []string{},
//				},
//			},
//			"health":  []health.Health{},
//			"metrics": []interface{}{},
//		})
//
//	batcher.Shutdown()
//	os.Unsetenv("STS_BATCHER_CAPACITY")
//}
//
//func TestBatcherStartSnapshot(t *testing.T) {
//	fwd := transactionforwarder.NewMockTransactionalForwarder()
//	batcher := newTransactionalBatcher(testHost, testAgent, 100, 15*time.Second)
//
//	batcher.SubmitStartSnapshot(testID, testTransactionID, testInstance)
//	batcher.SubmitComplete(testID)
//
//	message, _ := fwd.NextIntakePayload()
//
//	assert.Equal(t, message,
//		map[string]interface{}{
//			"internalHostname": "myhost",
//			"topologies": []topology.Topology{
//				{
//					StartSnapshot: true,
//					StopSnapshot:  false,
//					Instance:      testInstance,
//					Components:    []topology.Component{},
//					Relations:     []topology.Relation{},
//					DeleteIDs:     []string{},
//				},
//			},
//			"health":  []health.Health{},
//			"metrics": []interface{}{},
//		})
//
//	batcher.Shutdown()
//}
//
//func TestBatcherRelation(t *testing.T) {
//	fwd := transactionforwarder.NewMockTransactionalForwarder()
//	batcher := newTransactionalBatcher(testHost, testAgent, 100, 15*time.Second)
//
//	batcher.SubmitRelation(testID, testTransactionID, testInstance, testRelation)
//	batcher.SubmitComplete(testID)
//
//	message, _ := fwd.NextIntakePayload()
//
//	assert.Equal(t, message,
//		map[string]interface{}{
//			"internalHostname": "myhost",
//			"topologies": []topology.Topology{
//				{
//					StartSnapshot: false,
//					StopSnapshot:  false,
//					Instance:      testInstance,
//					Components:    []topology.Component{},
//					Relations:     []topology.Relation{testRelation},
//					DeleteIDs:     []string{},
//				},
//			},
//			"health":  []health.Health{},
//			"metrics": []interface{}{},
//		})
//
//	batcher.Shutdown()
//}
//
//func TestBatcherHealthStartSnapshot(t *testing.T) {
//	fwd := transactionforwarder.NewMockTransactionalForwarder()
//	batcher := newTransactionalBatcher(testHost, testAgent, 100, 15*time.Second)
//
//	batcher.SubmitHealthStartSnapshot(testID, testTransactionID, testStream, 1, 0)
//	batcher.SubmitComplete(testID)
//
//	message, _ := fwd.NextIntakePayload()
//
//	assert.Equal(t, message,
//		map[string]interface{}{
//			"internalHostname": "myhost",
//			"topologies":       []topology.Topology{},
//			"health": []health.Health{
//				{
//					StartSnapshot: testStartSnapshot,
//					Stream:        testStream,
//					CheckStates:   []health.CheckData{},
//				},
//			},
//			"metrics": []interface{}{},
//		})
//
//	batcher.Shutdown()
//}
//
//func TestBatchMultipleHealthStreams(t *testing.T) {
//	fwd := transactionforwarder.NewMockTransactionalForwarder()
//	batcher := newTransactionalBatcher(testHost, testAgent, 100, 15*time.Second)
//
//	batcher.SubmitHealthStartSnapshot(testID, testTransactionID, testStream, 1, 0)
//	batcher.SubmitHealthStartSnapshot(testID, testTransactionID, testStream2, 1, 0)
//	batcher.SubmitComplete(testID)
//
//	message, _ := fwd.NextIntakePayload()
//
//	assert.ObjectsAreEqualValues(message, map[string]interface{}{
//		"internalHostname": "myhost",
//		"topologies":       []topology.Topology{},
//		"health": []health.Health{
//			{
//				StartSnapshot: testStartSnapshot,
//				Stream:        testStream,
//				CheckStates:   []health.CheckData{},
//			},
//			{
//				StartSnapshot: testStartSnapshot,
//				Stream:        testStream2,
//				CheckStates:   []health.CheckData{},
//			},
//		},
//		"metrics": []interface{}{},
//	})
//
//	batcher.Shutdown()
//}
