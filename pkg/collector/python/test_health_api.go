//go:build python && test

package python

import (
	"encoding/json"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check/handler"
	"github.com/DataDog/datadog-agent/pkg/collector/check/test"
	check2 "github.com/StackVista/stackstate-receiver-go-client/pkg/model/check"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/health"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/transactional/transactionbatcher"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// #include <datadog_agent_rtloader.h>
import "C"

var expectedCheckData = health.CheckData{
	Unstructured: map[string]interface{}{
		"key":        "value ®",
		"stringlist": []interface{}{"a", "b", "c"},
		"boollist":   []interface{}{true, false},
		"intlist":    []interface{}{float64(1)},
		"doublelist": []interface{}{0.7, 1.42},
		"emptykey":   nil,
		"nestedobject": map[string]interface{}{
			"nestedkey": "nestedValue",
			"animals": map[string]interface{}{
				"legs":  "dog",
				"wings": "eagle",
				"tail":  "crocodile",
			},
		},
	},
}

func testHealthCheckData(t *testing.T) {
	_, mockTransactionalBatcher, _, checkManager := handler.SetupMockTransactionalComponents()

	testCheck := &test.STSTestCheck{Name: "check-id-health-check-data"}
	checkManager.RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	c := &health.Payload{
		Stream: health.Stream{
			Urn:       "myurn",
			SubStream: "substream",
		},
		Data: expectedCheckData,
	}
	data, err := json.Marshal(c)
	assert.NoError(t, err)

	checkId := C.CString(testCheck.String())
	stream := C.health_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("substream")
	StartTransaction(checkId)
	SubmitHealthStartSnapshot(checkId, &stream, C.int(0), C.int(1))
	SubmitHealthCheckData(
		checkId,
		&stream,
		C.CString(string(data)))
	SubmitHealthStopSnapshot(checkId, &stream)

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	actualTopology, found := mockTransactionalBatcher.GetCheckState(check2.CheckID(testCheck.ID()))
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())
	expectedStream := health.Stream{Urn: "myurn", SubStream: "substream"}

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: actualTopology.Transaction, // not asserting this specifically, it just needs to be present
		Health: map[string]health.Health{
			expectedStream.GoString(): {
				StartSnapshot: &health.StartSnapshotMetadata{RepeatIntervalS: 1, ExpiryIntervalS: 0},
				StopSnapshot:  &health.StopSnapshotMetadata{},
				Stream:        expectedStream,
				CheckStates: []health.CheckData{
					expectedCheckData,
				},
			},
		},
	}, actualTopology)

	checkManager.UnsubscribeCheckHandler(testCheck.ID())
}

func testHealthStartSnapshot(t *testing.T) {
	_, mockTransactionalBatcher, _, checkManager := handler.SetupMockTransactionalComponents()

	testCheck := &test.STSTestCheck{Name: "check-id-health-start-snapshot"}
	checkManager.RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	checkId := C.CString(testCheck.String())
	stream := C.health_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("substream")

	StartTransaction(checkId)
	SubmitHealthStartSnapshot(checkId, &stream, C.int(0), C.int(1))

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	actualTopology, found := mockTransactionalBatcher.GetCheckState(check2.CheckID(testCheck.ID()))
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())
	expectedStream := health.Stream{Urn: "myurn", SubStream: "substream"}

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: actualTopology.Transaction, // not asserting this specifically, it just needs to be present
		Health: map[string]health.Health{
			expectedStream.GoString(): {
				StartSnapshot: &health.StartSnapshotMetadata{RepeatIntervalS: 1, ExpiryIntervalS: 0},
				Stream:        expectedStream,
				CheckStates:   []health.CheckData{},
			},
		},
	}, actualTopology)

	checkManager.UnsubscribeCheckHandler(testCheck.ID())
}

func testHealthStopSnapshot(t *testing.T) {
	_, mockTransactionalBatcher, _, checkManager := handler.SetupMockTransactionalComponents()

	testCheck := &test.STSTestCheck{Name: "check-id-health-stop-snapshot"}
	checkManager.RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	checkId := C.CString(testCheck.String())
	stream := C.health_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("substream")
	StartTransaction(checkId)
	SubmitHealthStopSnapshot(checkId, &stream)

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	actualTopology, found := mockTransactionalBatcher.GetCheckState(check2.CheckID(testCheck.ID()))
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())
	expectedStream := health.Stream{Urn: "myurn", SubStream: "substream"}

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: actualTopology.Transaction, // not asserting this specifically, it just needs to be present
		Health: map[string]health.Health{
			expectedStream.GoString(): {
				StopSnapshot: &health.StopSnapshotMetadata{},
				Stream:       expectedStream,
				CheckStates:  []health.CheckData{},
			},
		},
	}, actualTopology)

	checkManager.UnsubscribeCheckHandler(testCheck.ID())
}

func testNoSubStream(t *testing.T) {
	_, mockTransactionalBatcher, _, checkManager := handler.SetupMockTransactionalComponents()

	testCheck := &test.STSTestCheck{Name: "check-id-health-no-sub-stream"}
	checkManager.RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	checkId := C.CString(testCheck.String())
	stream := C.health_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("")

	StartTransaction(checkId)
	SubmitHealthStartSnapshot(checkId, &stream, C.int(0), C.int(1))

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	actualTopology, found := mockTransactionalBatcher.GetCheckState(check2.CheckID(testCheck.ID()))
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())
	expectedStream := health.Stream{Urn: "myurn"}

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: actualTopology.Transaction, // not asserting this specifically, it just needs to be present
		Health: map[string]health.Health{
			expectedStream.GoString(): {
				StartSnapshot: &health.StartSnapshotMetadata{RepeatIntervalS: 1, ExpiryIntervalS: 0},
				Stream:        expectedStream,
				CheckStates:   []health.CheckData{},
			},
		},
	}, actualTopology)

	checkManager.UnsubscribeCheckHandler(testCheck.ID())
}
