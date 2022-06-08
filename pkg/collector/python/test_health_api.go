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
	testCheck := &check.STSTestCheck{Name: "check-id-health-check-data"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()
	mockTransactionalManager := transactionmanager.NewMockTransactionManager()

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

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
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

	checkmanager.GetCheckManager().Stop()
	mockTransactionalBatcher.Stop()
	mockTransactionalManager.Stop()
}

func testHealthStartSnapshot(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	testCheck := &check.STSTestCheck{Name: "check-id-health-start-snapshot"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()
	mockTransactionalManager := transactionmanager.NewMockTransactionManager()

	checkId := C.CString(testCheck.String())
	stream := C.health_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("substream")

	StartTransaction(checkId)
	SubmitHealthStartSnapshot(checkId, &stream, C.int(0), C.int(1))

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
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

	checkmanager.GetCheckManager().Stop()
	mockTransactionalBatcher.Stop()
	mockTransactionalManager.Stop()
}

func testHealthStopSnapshot(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	testCheck := &check.STSTestCheck{Name: "check-id-health-stop-snapshot"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()
	mockTransactionalManager := transactionmanager.NewMockTransactionManager()

	checkId := C.CString(testCheck.String())
	stream := C.health_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("substream")
	StartTransaction(checkId)
	SubmitHealthStopSnapshot(checkId, &stream)

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
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

	checkmanager.GetCheckManager().Stop()
	mockTransactionalBatcher.Stop()
	mockTransactionalManager.Stop()
}

func testNoSubStream(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	testCheck := &check.STSTestCheck{Name: "check-id-health-no-sub-stream"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()
	mockTransactionalManager := transactionmanager.NewMockTransactionManager()

	checkId := C.CString(testCheck.String())
	stream := C.health_stream_t{}
	stream.urn = C.CString("myurn")
	stream.sub_stream = C.CString("")

	StartTransaction(checkId)
	SubmitHealthStartSnapshot(checkId, &stream, C.int(0), C.int(1))

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
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

	checkmanager.GetCheckManager().Stop()
	mockTransactionalBatcher.Stop()
	mockTransactionalManager.Stop()
}
