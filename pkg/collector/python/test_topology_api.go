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
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// #include <datadog_agent_rtloader.h>
import "C"

var (
	mockTransactionalBatcher = transactionbatcher.NewMockTransactionalBatcher()
	mockTransactionManager   = transactionmanager.NewMockTransactionManager()
)

func init() {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})

}

func testComponentTopology(t *testing.T) {
	testCheck := &check.STSTestCheck{Name: "check-id-component-test"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	c := &topology.Component{
		ExternalID: "external-id",
		Type:       topology.Type{Name: "component-type"},
		Data: map[string]interface{}{
			"some": "data",
		},
	}
	data, err := json.Marshal(c)
	assert.NoError(t, err)

	checkId := C.CString(string(testCheck.ID()))
	instanceKey := C.instance_key_t{}
	instanceKey.type_ = C.CString("instance-type")
	instanceKey.url = C.CString("instance-url")
	SubmitStartSnapshot(checkId, &instanceKey)
	SubmitComponent(
		checkId,
		&instanceKey,
		C.CString("external-id"),
		C.CString("component-type"),
		C.CString(string(data)))
	SubmitStopSnapshot(checkId, &instanceKey)

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())
	expectedTopology := transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: &transactionbatcher.BatchTransaction{
			TransactionID:        "", // no start transaction, so the transaction is empty in this case
			CompletedTransaction: false,
		},
		Topology: &topology.Topology{
			StartSnapshot: true,
			StopSnapshot:  true,
			Instance:      topology.Instance{Type: "instance-type", URL: "instance-url"},
			Components: []topology.Component{
				{
					ExternalID: "external-id",
					Type:       topology.Type{Name: "component-type"},
					Data:       topology.Data{"some": "data"},
				},
			},
			Relations: []topology.Relation{},
			DeleteIDs: []string{},
		},
		Health: map[string]health.Health{},
	}
	assert.Equal(t, expectedTopology, actualTopology)

	checkmanager.GetCheckManager().UnsubscribeCheckHandler(testCheck.ID())
}

func testRelationTopology(t *testing.T) {
	testCheck := &check.STSTestCheck{Name: "check-id-relation-test"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	c := &topology.Relation{
		SourceID: "source-id",
		TargetID: "target-id",
		Type:     topology.Type{Name: "relation-type"},
		Data: map[string]interface{}{
			"some": "data",
		},
	}
	data, err := json.Marshal(c)
	assert.NoError(t, err)

	checkId := C.CString(string(testCheck.ID()))
	instanceKey := C.instance_key_t{}
	instanceKey.type_ = C.CString("instance-type")
	instanceKey.url = C.CString("instance-url")
	SubmitRelation(
		checkId,
		&instanceKey,
		C.CString("source-id"),
		C.CString("target-id"),
		C.CString("relation-type"),
		C.CString(string(data)))

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())
	expectedTopology := transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: &transactionbatcher.BatchTransaction{
			TransactionID:        "", // no start transaction, so the transaction is empty in this case
			CompletedTransaction: false,
		},
		Topology: &topology.Topology{
			StartSnapshot: false,
			StopSnapshot:  false,
			Instance:      topology.Instance{Type: "instance-type", URL: "instance-url"},
			Components:    []topology.Component{},
			Relations: []topology.Relation{
				{
					ExternalID: "source-id-relation-type-target-id",
					SourceID:   "source-id",
					TargetID:   "target-id",
					Type:       topology.Type{Name: "relation-type"},
					Data:       topology.Data{"some": "data"},
				},
			},
			DeleteIDs: []string{},
		},
		Health: map[string]health.Health{},
	}
	assert.Equal(t, expectedTopology, actualTopology)

	checkmanager.GetCheckManager().UnsubscribeCheckHandler(testCheck.ID())
}

func testStartTransaction(t *testing.T) {
	testCheck := &check.STSTestCheck{Name: "check-id-start-transaction"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})
	checkId := C.CString(string(testCheck.ID()))

	SubmitStartTransaction(checkId)
	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	transactionID := mockTransactionManager.GetCurrentTransaction()
	assert.NotEmpty(t, transactionID)
}

func testStopTransaction(t *testing.T) {
	testCheck := &check.STSTestCheck{Name: "check-id-stop-transaction"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})
	checkId := C.CString(string(testCheck.ID()))

	SubmitStartTransaction(checkId)
	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	transactionID := mockTransactionManager.GetCurrentTransaction()

	SubmitStopTransaction(checkId)

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: &transactionbatcher.BatchTransaction{
			TransactionID:        transactionID,
			CompletedTransaction: true,
		},
		Health: map[string]health.Health{},
	}, actualTopology)

	checkmanager.GetCheckManager().UnsubscribeCheckHandler(testCheck.ID())
}

func testStartSnapshotCheck(t *testing.T) {
	testCheck := &check.STSTestCheck{Name: "check-id-start-snapshot"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	checkId := C.CString(string(testCheck.ID()))
	instanceKey := C.instance_key_t{}
	instanceKey.type_ = C.CString("instance-type")
	instanceKey.url = C.CString("instance-url")
	SubmitStartSnapshot(checkId, &instanceKey)

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: &transactionbatcher.BatchTransaction{
			TransactionID:        "", // no start transaction, so the transaction is empty in this case
			CompletedTransaction: false,
		},
		Topology: &topology.Topology{
			StartSnapshot: true,
			StopSnapshot:  false,
			Instance:      topology.Instance{Type: "instance-type", URL: "instance-url"},
			Components:    []topology.Component{},
			Relations:     []topology.Relation{},
			DeleteIDs:     []string{},
		},
		Health: map[string]health.Health{},
	}, actualTopology)

	checkmanager.GetCheckManager().UnsubscribeCheckHandler(testCheck.ID())
}

func testStopSnapshotCheck(t *testing.T) {
	testCheck := &check.STSTestCheck{Name: "check-id-stop-snapshot"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	checkId := C.CString(string(testCheck.ID()))
	instanceKey := C.instance_key_t{}
	instanceKey.type_ = C.CString("instance-type")
	instanceKey.url = C.CString("instance-url")
	SubmitStopSnapshot(checkId, &instanceKey)

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())

	expectedTopology := transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: &transactionbatcher.BatchTransaction{
			TransactionID:        "", // no start transaction, so the transaction is empty in this case
			CompletedTransaction: false,
		},
		Topology: &topology.Topology{
			StartSnapshot: false,
			StopSnapshot:  true,
			Instance:      topology.Instance{Type: "instance-type", URL: "instance-url"},
			Components:    []topology.Component{},
			Relations:     []topology.Relation{},
			DeleteIDs:     []string{},
		},
		Health: map[string]health.Health{},
	}

	assert.Equal(t, expectedTopology, actualTopology)

	checkmanager.GetCheckManager().UnsubscribeCheckHandler(testCheck.ID())
}

func testDeleteTopologyElement(t *testing.T) {
	testCheck := &check.STSTestCheck{Name: "check-id-delete-element"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	checkID := C.CString(string(testCheck.ID()))
	instanceKey := C.instance_key_t{}
	instanceKey.type_ = C.CString("instance-type")
	instanceKey.url = C.CString("instance-url")
	topoElementId := "topo-element-id"

	SubmitStartSnapshot(checkID, &instanceKey)
	SubmitDelete(
		checkID,
		&instanceKey,
		C.CString(topoElementId))
	SubmitStopSnapshot(checkID, &instanceKey)

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())
	expectedTopology := transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: &transactionbatcher.BatchTransaction{
			TransactionID:        "", // no start transaction, so the transaction is empty in this case
			CompletedTransaction: false,
		},
		Topology: &topology.Topology{
			StartSnapshot: true,
			StopSnapshot:  true,
			Instance:      topology.Instance{Type: "instance-type", URL: "instance-url"},
			Components:    []topology.Component{},
			Relations:     []topology.Relation{},
			DeleteIDs:     []string{topoElementId},
		},
		Health: map[string]health.Health{},
	}
	assert.Equal(t, expectedTopology, actualTopology)

	checkmanager.GetCheckManager().UnsubscribeCheckHandler(testCheck.ID())
}
