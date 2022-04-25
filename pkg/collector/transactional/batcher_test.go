package transactional

import (
	"encoding/json"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/StackVista/stackstate-agent/pkg/telemetry"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

var (
	testInstance  = topology.Instance{Type: "mytype", URL: "myurl"}
	testInstance2 = topology.Instance{Type: "mytype2", URL: "myurl2"}
	testHost      = "myhost"
	testAgent     = "myagent"
	testID        = check.ID("myid")
	testID2       = check.ID("myid2")
	testComponent = topology.Component{
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

	testRawMetricsDataIntakeMetric  = testRawMetricsData.ConvertToIntakeMetric()
	testRawMetricsDataIntakeMetric2 = testRawMetricsData2.ConvertToIntakeMetric()
)

type MockForwarder struct {
	t               *testing.T
	ExpectedPayload TransactionalPayload
	DoneChannel     chan<- bool
}

func (tf MockForwarder) SubmitTransactionalIntake(payload TransactionalPayload) {
	assert.Equal(tf.t, tf.ExpectedPayload.transactionID, payload.transactionID)
	assert.Equal(tf.t, tf.ExpectedPayload.actionID, payload.actionID)

	message := map[string]interface{}{}
	err := json.Unmarshal(payload.payload, &message)
	assert.NoError(tf.t, err)

	assert.Equal(tf.t, tf.ExpectedPayload.payload, message)

	tf.DoneChannel <- true
}

func makeTransactionalPayload(t *testing.T, data map[string]interface{}, transactionID, actionID string) TransactionalPayload {
	payload, err := json.Marshal(data)
	assert.NoError(t, err)

	return TransactionalPayload{
		payload:       payload,
		transactionID: transactionID,
		actionID:      actionID,
	}
}

func TestBatchFlushOnStopSnapshot(t *testing.T) {
	data := map[string]interface{}{
		"internalHostname": "myhost",
		"topologies": []topology.Topology{
			{
				StartSnapshot: false,
				StopSnapshot:  true,
				Instance:      testInstance,
				Components:    []topology.Component{},
				Relations:     []topology.Relation{},
				DeleteIDs:     []string{},
			},
		},
		"health":  []health.Health{},
		"metrics": []interface{}{},
	}
	doneChannel := make(chan bool, 1)
	fwd := MockForwarder{
		t:               t,
		ExpectedPayload: makeTransactionalPayload(t, data, "", ""),
		DoneChannel:     doneChannel,
	}

	batcher := MakeCheckInstanceBatcher("test", testHost, testAgent, 100, 10*time.Second, fwd)

	batcher.SubmitStopSnapshot(testID, testInstance)

	_ = <-doneChannel
	batcher.Shutdown()
}

func TestBatchFlushOnStopHealthSnapshot(t *testing.T) {
	data := map[string]interface{}{
		"internalHostname": "myhost",
		"topologies":       []topology.Topology{},
		"health": []health.Health{
			{
				StopSnapshot: testStopSnapshot,
				Stream:       testStream,
				CheckStates:  []health.CheckData{},
			},
		},
		"metrics": []interface{}{},
	}

	doneChannel := make(chan bool, 1)
	fwd := MockForwarder{
		t:               t,
		ExpectedPayload: makeTransactionalPayload(t, data, "", ""),
		DoneChannel:     doneChannel,
	}

	batcher := MakeCheckInstanceBatcher("test", testHost, testAgent, 100, 10*time.Second, fwd)

	batcher.SubmitHealthStopSnapshot(testID, testStream)

	_ = <-doneChannel
	batcher.Shutdown()
}

func TestBatchFlushOnComplete(t *testing.T) {
	data := map[string]interface{}{
		"internalHostname": "myhost",
		"topologies": []topology.Topology{
			{
				StartSnapshot: false,
				StopSnapshot:  false,
				Instance:      testInstance,
				Components:    []topology.Component{testComponent},
				Relations:     []topology.Relation{},
				DeleteIDs:     []string{},
			},
		},
		"health": []health.Health{
			{
				Stream:      testStream,
				CheckStates: []health.CheckData{testCheckData},
			},
		},
		"metrics": []interface{}{testRawMetricsDataIntakeMetric, testRawMetricsDataIntakeMetric2},
	}
	doneChannel := make(chan bool, 1)
	fwd := MockForwarder{
		t:               t,
		ExpectedPayload: makeTransactionalPayload(t, data, "", ""),
		DoneChannel:     doneChannel,
	}

	batcher := MakeCheckInstanceBatcher("test", testHost, testAgent, 100, 10*time.Second, fwd)

	batcher.SubmitComponent(testID, testInstance, testComponent)
	batcher.SubmitHealthCheckData(testID, testStream, testCheckData)
	batcher.SubmitRawMetricsData(testID, testRawMetricsData)
	batcher.SubmitRawMetricsData(testID, testRawMetricsData2)
	batcher.SubmitComplete(testID)

	_ = <-doneChannel
	batcher.Shutdown()
}

func TestBatchNoDataNoComplete(t *testing.T) {
	data := map[string]interface{}{
		"internalHostname": "myhost",
		"topologies": []topology.Topology{
			{
				StartSnapshot: false,
				StopSnapshot:  true,
				Instance:      testInstance,
				Components:    []topology.Component{testComponent},
				Relations:     []topology.Relation{},
				DeleteIDs:     []string{},
			},
		},
		"health":  []health.Health{},
		"metrics": []interface{}{},
	}
	doneChannel := make(chan bool, 1)
	fwd := MockForwarder{
		t:               t,
		ExpectedPayload: makeTransactionalPayload(t, data, "", ""),
		DoneChannel:     doneChannel,
	}

	batcher := MakeCheckInstanceBatcher("test", testHost, testAgent, 100, 10*time.Second, fwd)

	batcher.SubmitComponent(testID, testInstance, testComponent)

	batcher.SubmitComplete(testID2)

	// We now send a stop to trigger a combined commit
	batcher.SubmitStopSnapshot(testID, testInstance)

	_ = <-doneChannel
	batcher.Shutdown()
}

func TestBatchMultipleTopologiesAndHealthStreams(t *testing.T) {
	data := map[string]interface{}{
		"internalHostname": "myhost",
		"topologies": []topology.Topology{
			{
				StartSnapshot: true,
				StopSnapshot:  true,
				Instance:      testInstance,
				Components:    []topology.Component{testComponent},
				Relations:     []topology.Relation{},
			},
			{
				StartSnapshot: false,
				StopSnapshot:  false,
				Instance:      testInstance2,
				Components:    []topology.Component{testComponent, testComponent, testComponent},
				Relations:     []topology.Relation{},
			},
		},
		"health": []health.Health{
			{
				StartSnapshot: testStartSnapshot,
				Stream:        testStream,
				CheckStates:   []health.CheckData{testCheckData},
			},
			{
				Stream:      testStream2,
				CheckStates: []health.CheckData{testCheckData},
			},
		},
		"metrics": []interface{}{
			testRawMetricsDataIntakeMetric,
			testRawMetricsDataIntakeMetric,
			testRawMetricsDataIntakeMetric2,
			testRawMetricsDataIntakeMetric2,
		},
	}
	doneChannel := make(chan bool, 1)
	fwd := MockForwarder{
		t:               t,
		ExpectedPayload: makeTransactionalPayload(t, data, "", ""),
		DoneChannel:     doneChannel,
	}

	batcher := MakeCheckInstanceBatcher("test", testHost, testAgent, 100, 10*time.Second, fwd)

	batcher.SubmitStartSnapshot(testID, testInstance)
	batcher.SubmitComponent(testID, testInstance, testComponent)
	batcher.SubmitComponent(testID2, testInstance2, testComponent)
	batcher.SubmitComponent(testID2, testInstance2, testComponent)
	batcher.SubmitComponent(testID2, testInstance2, testComponent)

	batcher.SubmitHealthStartSnapshot(testID, testStream, 1, 0)
	batcher.SubmitHealthCheckData(testID, testStream, testCheckData)
	batcher.SubmitHealthCheckData(testID2, testStream2, testCheckData)

	batcher.SubmitRawMetricsData(testID, testRawMetricsData)
	batcher.SubmitRawMetricsData(testID2, testRawMetricsData)
	batcher.SubmitRawMetricsData(testID, testRawMetricsData2)
	batcher.SubmitRawMetricsData(testID2, testRawMetricsData2)

	batcher.SubmitStopSnapshot(testID, testInstance)

	_ = <-doneChannel
	batcher.Shutdown()
}

func TestBatchFlushOnMaxElements(t *testing.T) {
	data := map[string]interface{}{
		"internalHostname": "myhost",
		"topologies": []topology.Topology{
			{
				StartSnapshot: false,
				StopSnapshot:  false,
				Instance:      testInstance,
				Components:    []topology.Component{testComponent, testComponent2},
				Relations:     []topology.Relation{},
				DeleteIDs:     []string{},
			},
		},
		"health":  []health.Health{},
		"metrics": []interface{}{},
	}

	doneChannel := make(chan bool, 1)
	fwd := MockForwarder{
		t:               t,
		ExpectedPayload: makeTransactionalPayload(t, data, "", ""),
		DoneChannel:     doneChannel,
	}

	batcher := MakeCheckInstanceBatcher("test", testHost, testAgent, 100, 10*time.Second, fwd)

	batcher.SubmitComponent(testID, testInstance, testComponent)
	batcher.SubmitComponent(testID, testInstance, testComponent2)

	_ = <-doneChannel
	batcher.Shutdown()
}

func TestBatchFlushOnMaxHealthElements(t *testing.T) {
	data := map[string]interface{}{
		"internalHostname": "myhost",
		"topologies":       []topology.Topology{},
		"health": []health.Health{
			{
				Stream:      testStream,
				CheckStates: []health.CheckData{testCheckData, testCheckData},
			},
		},
		"metrics": []interface{}{},
	}
	doneChannel := make(chan bool, 1)
	fwd := MockForwarder{
		t:               t,
		ExpectedPayload: makeTransactionalPayload(t, data, "", ""),
		DoneChannel:     doneChannel,
	}

	batcher := MakeCheckInstanceBatcher("test", testHost, testAgent, 100, 10*time.Second, fwd)

	batcher.SubmitHealthCheckData(testID, testStream, testCheckData)
	batcher.SubmitHealthCheckData(testID, testStream, testCheckData)

	_ = <-doneChannel
	batcher.Shutdown()
}

func TestBatchFlushOnMaxRawMetricsElements(t *testing.T) {
	data := map[string]interface{}{
		"internalHostname": "myhost",
		"topologies":       []topology.Topology{},
		"health":           []health.Health{},
		"metrics": []interface{}{
			testRawMetricsDataIntakeMetric, testRawMetricsDataIntakeMetric2},
	}
	doneChannel := make(chan bool, 1)
	fwd := MockForwarder{
		t:               t,
		ExpectedPayload: makeTransactionalPayload(t, data, "", ""),
		DoneChannel:     doneChannel,
	}

	batcher := MakeCheckInstanceBatcher("test", testHost, testAgent, 100, 10*time.Second, fwd)

	batcher.SubmitRawMetricsData(testID, testRawMetricsData)
	batcher.SubmitRawMetricsData(testID, testRawMetricsData2)

	_ = <-doneChannel
	batcher.Shutdown()
}

func TestBatchFlushOnMaxElementsEnv(t *testing.T) {
	data := map[string]interface{}{
		"internalHostname": "myhost",
		"topologies": []topology.Topology{
			{
				StartSnapshot: false,
				StopSnapshot:  false,
				Instance:      testInstance,
				Components:    []topology.Component{testComponent},
				Relations:     []topology.Relation{},
				DeleteIDs:     []string{},
			},
		},
		"health":  []health.Health{},
		"metrics": []interface{}{},
	}
	doneChannel := make(chan bool, 1)
	fwd := MockForwarder{
		t:               t,
		ExpectedPayload: makeTransactionalPayload(t, data, "", ""),
		DoneChannel:     doneChannel,
	}

	// set batcher max capacity via ENV var
	os.Setenv("DD_BATCHER_CAPACITY", "1")
	batcher := MakeCheckInstanceBatcher("test", testHost, testAgent, 1, 100*time.Second, fwd)
	assert.Equal(t, 1, batcher.builder.maxCapacity)
	batcher.SubmitComponent(testID, testInstance, testComponent)

	_ = <-doneChannel
	batcher.Shutdown()
	os.Unsetenv("STS_BATCHER_CAPACITY")
}

func TestBatcherStartSnapshot(t *testing.T) {
	data := map[string]interface{}{
		"internalHostname": "myhost",
		"topologies": []topology.Topology{
			{
				StartSnapshot: true,
				StopSnapshot:  false,
				Instance:      testInstance,
				Components:    []topology.Component{},
				Relations:     []topology.Relation{},
				DeleteIDs:     []string{},
			},
		},
		"health":  []health.Health{},
		"metrics": []interface{}{},
	}
	doneChannel := make(chan bool, 1)
	fwd := MockForwarder{
		t:               t,
		ExpectedPayload: makeTransactionalPayload(t, data, "", ""),
		DoneChannel:     doneChannel,
	}

	batcher := MakeCheckInstanceBatcher("test", testHost, testAgent, 100, 10*time.Second, fwd)

	batcher.SubmitStartSnapshot(testID, testInstance)
	batcher.SubmitComplete(testID)

	_ = <-doneChannel
	batcher.Shutdown()
}

func TestBatcherRelation(t *testing.T) {
	data := map[string]interface{}{
		"internalHostname": "myhost",
		"topologies": []topology.Topology{
			{
				StartSnapshot: false,
				StopSnapshot:  false,
				Instance:      testInstance,
				Components:    []topology.Component{},
				Relations:     []topology.Relation{testRelation},
				DeleteIDs:     []string{},
			},
		},
		"health":  []health.Health{},
		"metrics": []interface{}{},
	}
	doneChannel := make(chan bool, 1)
	fwd := MockForwarder{
		t:               t,
		ExpectedPayload: makeTransactionalPayload(t, data, "", ""),
		DoneChannel:     doneChannel,
	}

	batcher := MakeCheckInstanceBatcher("test", testHost, testAgent, 100, 10*time.Second, fwd)

	batcher.SubmitRelation(testID, testInstance, testRelation)
	batcher.SubmitComplete(testID)

	_ = <-doneChannel
	batcher.Shutdown()
}

func TestBatcherHealthStartSnapshot(t *testing.T) {
	data := map[string]interface{}{
		"internalHostname": "myhost",
		"topologies":       []topology.Topology{},
		"health": []health.Health{
			{
				StartSnapshot: testStartSnapshot,
				Stream:        testStream,
				CheckStates:   []health.CheckData{},
			},
		},
		"metrics": []interface{}{},
	}
	doneChannel := make(chan bool, 1)
	fwd := MockForwarder{
		t:               t,
		ExpectedPayload: makeTransactionalPayload(t, data, "", ""),
		DoneChannel:     doneChannel,
	}

	batcher := MakeCheckInstanceBatcher("test", testHost, testAgent, 100, 10*time.Second, fwd)

	batcher.SubmitHealthStartSnapshot(testID, testStream, 1, 0)
	batcher.SubmitComplete(testID)

	_ = <-doneChannel
	batcher.Shutdown()
}

func TestBatchMultipleHealthStreams(t *testing.T) {
	data := map[string]interface{}{
		"internalHostname": "myhost",
		"topologies":       []topology.Topology{},
		"health": []health.Health{
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
		},
		"metrics": []interface{}{},
	}
	doneChannel := make(chan bool, 1)
	fwd := MockForwarder{
		t:               t,
		ExpectedPayload: makeTransactionalPayload(t, data, "", ""),
		DoneChannel:     doneChannel,
	}

	batcher := MakeCheckInstanceBatcher("test", testHost, testAgent, 100, 10*time.Second, fwd)

	batcher.SubmitHealthStartSnapshot(testID, testStream, 1, 0)
	batcher.SubmitHealthStartSnapshot(testID, testStream2, 1, 0)
	batcher.SubmitComplete(testID)

	_ = <-doneChannel
	batcher.Shutdown()
}
