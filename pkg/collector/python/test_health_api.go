//go:build python && test
// +build python,test

package python

import (
	"encoding/json"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/checkmanager"
	"github.com/StackVista/stackstate-agent/pkg/collector/check/handler"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/health"
	"github.com/stretchr/testify/assert"
	"testing"
)

// #include <datadog_agent_rtloader.h>
import "C"

var expectedCheckData = health.CheckData{
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
}

func testHealthCheckData(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	checkmanager.GetCheckManager().RegisterCheckHandler(&check.STSTestCheck{Name: "check-id"}, integration.Data{},
		integration.Data{})
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()
	transactionmanager.NewMockTransactionManager()

	c := &health.Payload{
		Stream: health.Stream{
			Urn:       "myurn",
			SubStream: "substream",
		},
		Data: expectedCheckData,
	}
	data, err := json.Marshal(c)
	assert.NoError(t, err)

	checkId := C.CString("check-id")
	stream := C.health_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("substream")
	SubmitHealthStartSnapshot(checkId, &stream, C.int(0), C.int(1))
	SubmitHealthCheckData(
		checkId,
		&stream,
		C.CString(string(data)))
	SubmitHealthStopSnapshot(checkId, &stream)

	expectedState := mockTransactionalBatcher.CollectedTopology.Flush()
	expectedStream := health.Stream{Urn: "myurn", SubStream: "substream"}

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
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
			},
		}), expectedState)
}

func testHealthStartSnapshot(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	checkmanager.GetCheckManager().RegisterCheckHandler(&check.STSTestCheck{Name: "check-id"}, integration.Data{},
		integration.Data{})
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()
	transactionmanager.NewMockTransactionManager()

	checkId := C.CString("check-id")
	stream := C.health_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("substream")
	SubmitHealthStartSnapshot(checkId, &stream, C.int(0), C.int(1))

	expectedState := mockTransactionalBatcher.CollectedTopology.Flush()
	expectedStream := health.Stream{Urn: "myurn", SubStream: "substream"}

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
				Health: map[string]health.Health{
					expectedStream.GoString(): {
						StartSnapshot: &health.StartSnapshotMetadata{RepeatIntervalS: 1, ExpiryIntervalS: 0},
						Stream:        expectedStream,
						CheckStates:   []health.CheckData{},
					},
				},
			},
		}), expectedState)
}

func testHealthStopSnapshot(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	checkmanager.GetCheckManager().RegisterCheckHandler(&check.STSTestCheck{Name: "check-id"}, integration.Data{},
		integration.Data{})
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()
	transactionmanager.NewMockTransactionManager()

	checkId := C.CString("check-id")
	stream := C.health_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("substream")
	SubmitHealthStopSnapshot(checkId, &stream)

	expectedState := mockTransactionalBatcher.CollectedTopology.Flush()
	expectedStream := health.Stream{Urn: "myurn", SubStream: "substream"}

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
				Health: map[string]health.Health{
					expectedStream.GoString(): {
						StopSnapshot: &health.StopSnapshotMetadata{},
						Stream:       expectedStream,
						CheckStates:  []health.CheckData{},
					},
				},
			},
		}), expectedState)
}

func testNoSubStream(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	checkmanager.GetCheckManager().RegisterCheckHandler(&check.STSTestCheck{Name: "check-id"}, integration.Data{},
		integration.Data{})
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()
	transactionmanager.NewMockTransactionManager()

	checkId := C.CString("check-id")
	stream := C.health_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("")
	SubmitHealthStartSnapshot(checkId, &stream, C.int(0), C.int(1))

	expectedState := mockTransactionalBatcher.CollectedTopology.Flush()
	expectedStream := health.Stream{Urn: "myurn"}

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
				Health: map[string]health.Health{
					expectedStream.GoString(): {
						StartSnapshot: &health.StartSnapshotMetadata{RepeatIntervalS: 1, ExpiryIntervalS: 0},
						Stream:        expectedStream,
						CheckStates:   []health.CheckData{},
					},
				},
			},
		}), expectedState)
}
