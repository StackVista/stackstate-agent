package transactionbatcher

import (
	"encoding/json"
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional/transactionforwarder"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional/transactionmanager"
	"github.com/DataDog/datadog-agent/pkg/config/setup"
	"github.com/DataDog/datadog-agent/pkg/health"
	"github.com/DataDog/datadog-agent/pkg/metrics/event"
	"github.com/DataDog/datadog-agent/pkg/telemetry"
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/stretchr/testify/assert"
	"os"
	"sort"
	"testing"
	"time"
)

var (
	testInstance       = topology.Instance{Type: "mytype", URL: "myurl"}
	testInstance2      = topology.Instance{Type: "mytype2", URL: "myurl2"}
	testHost           = "myhost"
	testAgent          = "myagent"
	testID             = checkid.ID("myid")
	testID2            = checkid.ID("myid2")
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
	testDeleteID1 = "delete-id-1"
	testDeleteID2 = "delete-id-2"

	testStream        = health.Stream{Urn: "urn", SubStream: "bla"}
	testStream2       = health.Stream{Urn: "urn"}
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
		HostName:  "Hostname",
		Value:     100,
		Tags: []string{
			"hello",
			"world",
		},
	}

	testRawMetricsDataIntakeMetric  = testRawMetricsData.IntakeMetricJSON()
	testRawMetricsDataIntakeMetric2 = testRawMetricsData2.IntakeMetricJSON()

	testEvent = event.Event{
		Title:          "test-event-1",
		Ts:             time.Now().Unix(),
		EventType:      "docker",
		Tags:           []string{"my", "test", "tags"},
		AggregationKey: "docker:redis",
		SourceTypeName: "docker",
		Priority:       event.EventPriorityNormal,
	}
	testEvent2 = event.Event{
		Title:          "test-event-2",
		Ts:             time.Now().Unix(),
		EventType:      "docker",
		Tags:           []string{"my", "test", "tags"},
		AggregationKey: "docker:mysql",
		SourceTypeName: "docker",
		Priority:       event.EventPriorityNormal,
		EventContext: &event.EventContext{
			Data:     map[string]interface{}{},
			Source:   "docker",
			Category: "containers",
		},
	}
	testEvent3 = event.Event{
		Title:          "test-event-3",
		Ts:             time.Now().Unix(),
		EventType:      "docker",
		Tags:           []string{"my", "test", "tags"},
		AggregationKey: "docker:mysql",
		SourceTypeName: "docker-other",
		Priority:       event.EventPriorityNormal,
		EventContext: &event.EventContext{
			Data:               map[string]interface{}{},
			Source:             "docker",
			Category:           "containers",
			ElementIdentifiers: []string{"element-identifier"},
			SourceLinks:        []event.SourceLink{{Title: "source-link", URL: "source-url"}},
		},
	}
)

func init() {
	transactionforwarder.NewMockTransactionalForwarder()
	transactionmanager.NewMockTransactionManager()
}

// TODO: these might hit nil pointers in the batcher because we only init the transaction manager and forwarder in the testBatcher function
// TODO: after the batcher operations have been executed
func testBatcher(t *testing.T, transactionState map[string]bool, expectedPayload transactional.IntakePayload) {
	tm := transactionmanager.GetTransactionManager().(*transactionmanager.MockTransactionManager)
	fwd := transactionforwarder.GetTransactionalForwarder().(*transactionforwarder.MockTransactionalForwarder)

	// get the action commit made by the batcher from the transaction manager for all the transactions in this payload
	commitActions := make(map[string]transactionmanager.CommitAction, len(transactionState))
	var foundTx []string
	for i := 0; i < len(transactionState); i++ {
		commitAction := tm.NextAction().(transactionmanager.CommitAction)
		_, found := transactionState[commitAction.TransactionID]
		if !found {
			assert.Fail(t, "Commit action for transaction %s, not found in expected transaction state: %v",
				commitAction.TransactionID, transactionState)
		}

		commitActions[commitAction.TransactionID] = commitAction
		foundTx = append(foundTx, commitAction.TransactionID)
	}

	// get the expected transactions in the transactionState
	var expectedTx []string
	for txID := range transactionState {
		expectedTx = append(expectedTx, txID)
	}

	// ensure that we found all transactions in the transactionmanager that we expected to be there
	sort.Strings(expectedTx)
	sort.Strings(foundTx)
	assert.Equal(t, expectedTx, foundTx)

	// get the intake payload that was produced for this action
	payload := fwd.NextPayload()
	actualPayload := transactional.NewIntakePayload()
	json.Unmarshal(payload.Body, &actualPayload)

	// assert the payload matches the expected payload for the data produced
	assert.Equal(t, expectedPayload.InternalHostname, actualPayload.InternalHostname)
	sort.Slice(actualPayload.Topologies, func(i, j int) bool {
		return actualPayload.Topologies[i].Instance.GoString() > actualPayload.Topologies[j].Instance.GoString()
	})
	assert.Equal(t, expectedPayload.Topologies, actualPayload.Topologies)
	sort.Slice(actualPayload.Health, func(i, j int) bool {
		return actualPayload.Health[i].Stream.GoString() < actualPayload.Health[j].Stream.GoString()
	})
	assert.Equal(t, expectedPayload.Health, actualPayload.Health)
	assert.Equal(t, expectedPayload.Metrics, actualPayload.Metrics)
	assert.Equal(t, len(expectedPayload.Events), len(actualPayload.Events))
	for key, expectedEvents := range expectedPayload.Events {
		actualEvents := actualPayload.Events[key]

		sort.Slice(actualEvents, func(i, j int) bool {
			return actualEvents[i].Title < actualEvents[j].Title
		})

		sort.Slice(expectedEvents, func(i, j int) bool {
			return expectedEvents[i].Title < expectedEvents[j].Title
		})

		for i, ev := range actualEvents {
			assert.Equal(t, expectedEvents[i].String(), ev.String())
		}
	}
	// assert the transaction map produced by the batcher contains the correct action id and completed status
	expectedTransactionMap := make(map[string]transactional.PayloadTransaction, len(commitActions))
	for i, ca := range commitActions {
		expectedTransactionMap[ca.TransactionID] = transactional.PayloadTransaction{
			ActionID:             ca.ActionID,
			CompletedTransaction: transactionState[i],
		}
	}

	assert.Equal(t, expectedTransactionMap, payload.TransactionActionMap)

}

func TestBatchNoPayloadOnlyCompleteTransaction(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100)
	batcher.SubmitCompleteTransaction(testID, testTransactionID)

	transactionStates := map[string]bool{
		testTransactionID: true,
	}
	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"

	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()
}

func TestBatchFlushSnapshotOnComplete(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100)
	batcher.SubmitStopSnapshot(testID, testTransactionID, testInstance)
	batcher.SubmitCompleteTransaction(testID, testTransactionID)

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

	transactionStates := map[string]bool{
		testTransactionID: true,
	}
	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()
}

func TestBatchFlushHealthOnComplete(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100)

	batcher.SubmitHealthStopSnapshot(testID, testTransactionID, testStream)
	batcher.SubmitCompleteTransaction(testID, testTransactionID)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Health = []health.Health{
		{
			StopSnapshot: testStopSnapshot,
			Stream:       testStream,
			CheckStates:  []health.CheckData{},
		},
	}

	transactionStates := map[string]bool{
		testTransactionID: true,
	}
	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()
}

func TestBatchFlushOnComplete(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100)

	batcher.SubmitComponent(testID, testTransactionID, testInstance, testComponent)
	batcher.SubmitHealthCheckData(testID, testTransactionID, testStream, testCheckData)
	batcher.SubmitRawMetricsData(testID, testTransactionID, testRawMetricsData)
	batcher.SubmitRawMetricsData(testID, testTransactionID, testRawMetricsData2)
	batcher.SubmitEvent(testID, testTransactionID, testEvent)
	batcher.SubmitCompleteTransaction(testID, testTransactionID)

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
	expectedPayload.Events = map[string][]event.Event{
		"docker": {testEvent},
	}

	transactionStates := map[string]bool{
		testTransactionID: true,
	}
	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()
}

func TestBatchDataCompleteTransaction(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100)

	batcher.StartTransaction(testID, testTransactionID)
	batcher.SubmitComponent(testID, testTransactionID, testInstance, testComponent)
	batcher.SubmitCompleteTransaction(testID2, testTransaction2ID)

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

	transactionStates := map[string]bool{
		testTransactionID:  false,
		testTransaction2ID: true,
	}
	testBatcher(t, transactionStates, expectedPayload)

	// We now send a stop to trigger a combined commit
	batcher.SubmitStopSnapshot(testID, testTransactionID, testInstance)
	batcher.SubmitCompleteTransaction(testID, testTransactionID)

	expectedPayload = transactional.NewIntakePayload()
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

	transactionStates = map[string]bool{
		testTransactionID: true,
	}
	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()
}

func TestBatchMultipleTopologiesAndHealthStreams(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100)

	batcher.SubmitStartSnapshot(testID, testTransactionID, testInstance)
	batcher.SubmitComponent(testID, testTransactionID, testInstance, testComponent)
	batcher.SubmitComponent(testID2, testTransaction2ID, testInstance2, testComponent)
	batcher.SubmitComponent(testID2, testTransaction2ID, testInstance2, testComponent)
	batcher.SubmitComponent(testID2, testTransaction2ID, testInstance2, testComponent)
	batcher.SubmitDelete(testID, testTransactionID, testInstance, testDeleteID1)
	batcher.SubmitDelete(testID2, testTransaction2ID, testInstance2, testDeleteID2)

	batcher.SubmitHealthStartSnapshot(testID, testTransactionID, testStream, 1, 0)
	batcher.SubmitHealthCheckData(testID, testTransactionID, testStream, testCheckData)
	batcher.SubmitHealthCheckData(testID2, testTransaction2ID, testStream2, testCheckData)

	batcher.SubmitRawMetricsData(testID, testTransactionID, testRawMetricsData)
	batcher.SubmitRawMetricsData(testID2, testTransaction2ID, testRawMetricsData)
	batcher.SubmitRawMetricsData(testID, testTransactionID, testRawMetricsData2)
	batcher.SubmitRawMetricsData(testID2, testTransaction2ID, testRawMetricsData2)

	batcher.SubmitEvent(testID, testTransactionID, testEvent)
	batcher.SubmitEvent(testID, testTransactionID, testEvent3)
	batcher.SubmitEvent(testID2, testTransaction2ID, testEvent2)

	batcher.SubmitStopSnapshot(testID, testTransactionID, testInstance)
	batcher.SubmitCompleteTransaction(testID, testTransactionID)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Topologies = []topology.Topology{
		{
			StartSnapshot: true,
			StopSnapshot:  true,
			Instance:      testInstance,
			Components:    []topology.Component{testComponent},
			Relations:     []topology.Relation{},
			DeleteIDs:     []string{testDeleteID1},
		},
		{
			StartSnapshot: false,
			StopSnapshot:  false,
			Instance:      testInstance2,
			Components:    []topology.Component{testComponent, testComponent, testComponent},
			Relations:     []topology.Relation{},
			DeleteIDs:     []string{testDeleteID2},
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
	// order in submission doesn't matter, each state (check) is added after one another
	expectedPayload.Metrics = []interface{}{
		testRawMetricsDataIntakeMetric,
		testRawMetricsDataIntakeMetric2,
		testRawMetricsDataIntakeMetric,
		testRawMetricsDataIntakeMetric2,
	}

	expectedPayload.Events = map[string][]event.Event{
		"docker":       {testEvent, testEvent2},
		"docker-other": {testEvent3},
	}

	transactionStates := map[string]bool{
		testTransactionID:  true,
		testTransaction2ID: false,
	}

	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()
}

func TestBatchFlushOnMaxElements(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 2)

	batcher.SubmitComponent(testID, testTransactionID, testInstance, testComponent)
	batcher.SubmitComponent(testID, testTransactionID, testInstance, testComponent2)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Topologies = []topology.Topology{
		{
			StartSnapshot: false,
			StopSnapshot:  false,
			Instance:      testInstance,
			Components:    []topology.Component{testComponent, testComponent2},
			Relations:     []topology.Relation{},
			DeleteIDs:     []string{},
		},
	}

	transactionStates := map[string]bool{
		testTransactionID: false,
	}

	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()
}

func TestBatchFlushOnMaxHealthElements(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 2)

	batcher.SubmitHealthCheckData(testID, testTransactionID, testStream, testCheckData)
	batcher.SubmitHealthCheckData(testID, testTransactionID, testStream, testCheckData)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Health = []health.Health{
		{
			Stream:      testStream,
			CheckStates: []health.CheckData{testCheckData, testCheckData},
		},
	}

	transactionStates := map[string]bool{
		testTransactionID: false,
	}

	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()
}

func TestBatchFlushOnMaxRawMetricsElements(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 2)

	batcher.SubmitRawMetricsData(testID, testTransactionID, testRawMetricsData)
	batcher.SubmitRawMetricsData(testID, testTransactionID, testRawMetricsData2)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Metrics = []interface{}{
		testRawMetricsDataIntakeMetric, testRawMetricsDataIntakeMetric2,
	}

	transactionStates := map[string]bool{
		testTransactionID: false,
	}

	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()
}

func TestBatchFlushOnMaxEvents(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 2)

	batcher.SubmitEvent(testID, testTransactionID, testEvent)
	batcher.SubmitEvent(testID, testTransactionID, testEvent2)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Events = map[string][]event.Event{
		"docker": {testEvent, testEvent2},
	}

	transactionStates := map[string]bool{
		testTransactionID: false,
	}

	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()
}

func TestBatchFlushOnMaxElementsEnv(t *testing.T) {
	// set transactionbatcher max capacity via ENV var
	os.Setenv("DD_BATCHER_CAPACITY", "1")
	batcher := newTransactionalBatcher(testHost, testAgent, setup.GetMaxCapacity())

	assert.Equal(t, 1, batcher.builder.maxCapacity)

	batcher.SubmitComponent(testID, testTransactionID, testInstance, testComponent)

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

	transactionStates := map[string]bool{
		testTransactionID: false,
	}

	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()

	os.Unsetenv("STS_BATCHER_CAPACITY")
}

func TestBatcherStartSnapshot(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100)

	batcher.SubmitStartSnapshot(testID, testTransactionID, testInstance)
	batcher.SubmitCompleteTransaction(testID, testTransactionID)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Topologies = []topology.Topology{
		{
			StartSnapshot: true,
			StopSnapshot:  false,
			Instance:      testInstance,
			Components:    []topology.Component{},
			Relations:     []topology.Relation{},
			DeleteIDs:     []string{},
		},
	}

	transactionStates := map[string]bool{
		testTransactionID: true,
	}

	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()
}

func TestBatcherRelation(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100)

	batcher.SubmitRelation(testID, testTransactionID, testInstance, testRelation)
	batcher.SubmitCompleteTransaction(testID, testTransactionID)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Topologies = []topology.Topology{
		{
			StartSnapshot: false,
			StopSnapshot:  false,
			Instance:      testInstance,
			Components:    []topology.Component{},
			Relations:     []topology.Relation{testRelation},
			DeleteIDs:     []string{},
		},
	}

	transactionStates := map[string]bool{
		testTransactionID: true,
	}

	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()
}

func TestBatcherHealthStartSnapshot(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100)

	batcher.SubmitHealthStartSnapshot(testID, testTransactionID, testStream, 1, 0)
	batcher.SubmitCompleteTransaction(testID, testTransactionID)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Health = []health.Health{
		{
			StartSnapshot: testStartSnapshot,
			Stream:        testStream,
			CheckStates:   []health.CheckData{},
		},
	}

	transactionStates := map[string]bool{
		testTransactionID: true,
	}

	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()
}

func TestBatchMultipleHealthStreams(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100)

	batcher.SubmitHealthStartSnapshot(testID, testTransactionID, testStream, 1, 0)
	batcher.SubmitHealthStartSnapshot(testID, testTransactionID, testStream2, 1, 0)
	batcher.SubmitCompleteTransaction(testID, testTransactionID)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Health = []health.Health{
		{
			StartSnapshot: testStartSnapshot,
			Stream:        testStream,
			CheckStates:   []health.CheckData{},
		},
		{
			StartSnapshot: testStartSnapshot,
			Stream:        testStream2,
			CheckStates:   []health.CheckData{},
		},
	}

	transactionStates := map[string]bool{
		testTransactionID: true,
	}

	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()
}

func TestBatchClearState(t *testing.T) {
	batcher := newTransactionalBatcher(testHost, testAgent, 100)

	batcher.StartTransaction(testID, testTransactionID)
	batcher.SubmitStartSnapshot(testID, testTransactionID, testInstance)
	batcher.SubmitComponent(testID, testTransactionID, testInstance, testComponent)
	batcher.SubmitDelete(testID, testTransactionID, testInstance, testDeleteID1)
	batcher.SubmitEvent(testID, testTransactionID, testEvent)

	// testID2 + testTransaction2ID will be cancelled and therefore should not be in the final payload
	batcher.StartTransaction(testID2, testTransaction2ID)
	batcher.SubmitStartSnapshot(testID2, testTransaction2ID, testInstance)
	batcher.SubmitComponent(testID2, testTransaction2ID, testInstance, testComponent)
	batcher.SubmitDelete(testID2, testTransaction2ID, testInstance, testDeleteID2)
	batcher.SubmitEvent(testID2, testTransaction2ID, testEvent2)
	batcher.SubmitClearState(testID2)

	batcher.SubmitCompleteTransaction(testID, testTransactionID)

	expectedPayload := transactional.NewIntakePayload()
	expectedPayload.InternalHostname = "myhost"
	expectedPayload.Topologies = []topology.Topology{
		{
			StartSnapshot: true,
			StopSnapshot:  false,
			Instance:      testInstance,
			Components:    []topology.Component{testComponent},
			Relations:     []topology.Relation{},
			DeleteIDs:     []string{testDeleteID1},
		},
	}
	expectedPayload.Events = map[string][]event.Event{
		"docker": {testEvent},
	}

	transactionStates := map[string]bool{
		testTransactionID: true,
	}

	testBatcher(t, transactionStates, expectedPayload)

	batcher.Stop()

}
