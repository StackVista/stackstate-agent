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

func testComponentTopology(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	checkmanager.GetCheckManager().RegisterCheckHandler(&check.STSTestCheck{Name: "check-id"}, integration.Data{},
		integration.Data{})
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()
	transactionmanager.NewMockTransactionManager()

	c := &topology.Component{
		ExternalID: "external-id",
		Type:       topology.Type{Name: "component-type"},
		Data: map[string]interface{}{
			"some": "data",
		},
	}
	data, err := json.Marshal(c)
	assert.NoError(t, err)

	checkId := C.CString("check-id")
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

	actualTopology := mockTransactionalBatcher.CollectedTopology.Flush()
	instance := topology.Instance{Type: "instance-type", URL: "instance-url"}

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
				Transaction: &transactionbatcher.BatchTransaction{
					TransactionID:        "", // no start transaction, so the transaction is empty in this case
					CompletedTransaction: false,
				},
				Topology: &topology.Topology{
					StartSnapshot: true,
					StopSnapshot:  true,
					Instance:      instance,
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
			},
		}), actualTopology)
}

func testRelationTopology(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	checkmanager.GetCheckManager().RegisterCheckHandler(&check.STSTestCheck{Name: "check-id"}, integration.Data{},
		integration.Data{})
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()
	transactionmanager.NewMockTransactionManager()

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

	checkId := C.CString("check-id")
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

	actualTopology := mockTransactionalBatcher.CollectedTopology.Flush()
	instance := topology.Instance{Type: "instance-type", URL: "instance-url"}

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
				Transaction: &transactionbatcher.BatchTransaction{
					TransactionID:        "", // no start transaction, so the transaction is empty in this case
					CompletedTransaction: false,
				},
				Topology: &topology.Topology{
					StartSnapshot: false,
					StopSnapshot:  false,
					Instance:      instance,
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
			},
		}), actualTopology)
}

func testStartTransaction(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	checkmanager.GetCheckManager().RegisterCheckHandler(&check.STSTestCheck{Name: "check-id"}, integration.Data{},
		integration.Data{})
	transactionManager := transactionmanager.NewMockTransactionManager()

	checkId := C.CString("check-id")

	SubmitStartTransaction(checkId)
	time.Sleep(100 * time.Millisecond) // sleep a bit for everything to complete

	transactionID := transactionManager.GetCurrentTransaction()
	assert.NotEmpty(t, transactionID)
}

func testStopTransaction(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	testCheck := &check.STSTestCheck{Name: "check-id"}
	checkmanager.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})
	transactionManager := transactionmanager.NewMockTransactionManager()
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()

	checkId := C.CString("check-id")

	SubmitStartTransaction(checkId)
	time.Sleep(100 * time.Millisecond) // sleep a bit for everything to complete

	transactionID := transactionManager.GetCurrentTransaction()

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

}

func testStartSnapshotCheck(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	checkmanager.GetCheckManager().RegisterCheckHandler(&check.STSTestCheck{Name: "check-id"}, integration.Data{},
		integration.Data{})
	transactionmanager.NewMockTransactionManager()
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()

	checkId := C.CString("check-id")
	instanceKey := C.instance_key_t{}
	instanceKey.type_ = C.CString("instance-type")
	instanceKey.url = C.CString("instance-url")
	SubmitStartSnapshot(checkId, &instanceKey)

	actualTopology := mockTransactionalBatcher.CollectedTopology.Flush()
	instance := topology.Instance{Type: "instance-type", URL: "instance-url"}

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
				Transaction: &transactionbatcher.BatchTransaction{
					TransactionID:        "", // no start transaction, so the transaction is empty in this case
					CompletedTransaction: false,
				},
				Topology: &topology.Topology{
					StartSnapshot: true,
					StopSnapshot:  false,
					Instance:      instance,
					Components:    []topology.Component{},
					Relations:     []topology.Relation{},
					DeleteIDs:     []string{},
				},
				Health: map[string]health.Health{},
			},
		}), actualTopology)
}

func testStopSnapshotCheck(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	checkmanager.GetCheckManager().RegisterCheckHandler(&check.STSTestCheck{Name: "check-id"}, integration.Data{},
		integration.Data{})
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()
	transactionmanager.NewMockTransactionManager()

	checkId := C.CString("check-id")
	instanceKey := C.instance_key_t{}
	instanceKey.type_ = C.CString("instance-type")
	instanceKey.url = C.CString("instance-url")
	SubmitStopSnapshot(checkId, &instanceKey)

	actualTopology := mockTransactionalBatcher.CollectedTopology.Flush()
	instance := topology.Instance{Type: "instance-type", URL: "instance-url"}

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
				Transaction: &transactionbatcher.BatchTransaction{
					TransactionID:        "", // no start transaction, so the transaction is empty in this case
					CompletedTransaction: false,
				},
				Topology: &topology.Topology{
					StartSnapshot: false,
					StopSnapshot:  true,
					Instance:      instance,
					Components:    []topology.Component{},
					Relations:     []topology.Relation{},
					DeleteIDs:     []string{},
				},
				Health: map[string]health.Health{},
			},
		}), actualTopology)
}

func testDeleteTopologyElement(t *testing.T) {
	checkmanager.InitCheckManager(handler.NoCheckReloader{})
	checkmanager.GetCheckManager().RegisterCheckHandler(&check.STSTestCheck{Name: "check-id"}, integration.Data{},
		integration.Data{})
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()
	transactionmanager.NewMockTransactionManager()

	checkID := C.CString("check-id")
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

	actualTopology := mockTransactionalBatcher.CollectedTopology.Flush()
	instance := topology.Instance{Type: "instance-type", URL: "instance-url"}

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
				Transaction: &transactionbatcher.BatchTransaction{
					TransactionID:        "", // no start transaction, so the transaction is empty in this case
					CompletedTransaction: false,
				},
				Topology: &topology.Topology{
					StartSnapshot: true,
					StopSnapshot:  true,
					Instance:      instance,
					Components:    []topology.Component{},
					Relations:     []topology.Relation{},
					DeleteIDs:     []string{topoElementId},
				},
				Health: map[string]health.Health{},
			},
		}), actualTopology)
}
