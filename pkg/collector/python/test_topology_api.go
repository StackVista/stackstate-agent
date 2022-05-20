//go:build python && test
// +build python,test

package python

import (
	"encoding/json"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/transactional/transactionmanager"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/stretchr/testify/assert"
	"testing"
)

// #include <datadog_agent_rtloader.h>
import "C"

func testComponentTopology(t *testing.T) {
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()

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

	expectedTopology := mockTransactionalBatcher.CollectedTopology.Flush()
	instance := topology.Instance{Type: "instance-type", URL: "instance-url"}

	assert.ObjectsAreEqualValues(expectedTopology, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
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
				},
			},
		}))
}

func testRelationTopology(t *testing.T) {
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()

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

	expectedTopology := mockTransactionalBatcher.CollectedTopology.Flush()
	instance := topology.Instance{Type: "instance-type", URL: "instance-url"}

	assert.ObjectsAreEqualValues(expectedTopology, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
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
				},
			},
		}))
}

func testStartTransaction(t *testing.T) {
	transactionManager := transactionmanager.NewMockTransactionManager()

	checkId := C.CString("check-id")

	SubmitStartTransaction(checkId)

	transactionID := transactionManager.GetCurrentTransaction()
	assert.NotEmpty(t, transactionID)
}

func testStopTransaction(t *testing.T) {
	transactionManager := transactionmanager.NewMockTransactionManager()
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()

	checkId := C.CString("check-id")

	SubmitStartTransaction(checkId)
	transactionID := transactionManager.GetCurrentTransaction()

	SubmitStopTransaction(checkId)

	expectedTopology := mockTransactionalBatcher.CollectedTopology.Flush()

	assert.ObjectsAreEqualValues(expectedTopology, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
				Transaction: &transactionbatcher.BatchTransaction{
					TransactionID:        transactionID,
					CompletedTransaction: true,
				},
			},
		}))
}

func testStartSnapshotCheck(t *testing.T) {
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()

	checkId := C.CString("check-id")
	instanceKey := C.instance_key_t{}
	instanceKey.type_ = C.CString("instance-type")
	instanceKey.url = C.CString("instance-url")
	SubmitStartSnapshot(checkId, &instanceKey)

	expectedTopology := mockTransactionalBatcher.CollectedTopology.Flush()
	instance := topology.Instance{Type: "instance-type", URL: "instance-url"}

	assert.ObjectsAreEqualValues(expectedTopology, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
				Topology: &topology.Topology{
					StartSnapshot: true,
					StopSnapshot:  false,
					Instance:      instance,
					Components:    []topology.Component{},
					Relations:     []topology.Relation{},
				},
			},
		}))
}

func testStopSnapshotCheck(t *testing.T) {
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()

	checkId := C.CString("check-id")
	instanceKey := C.instance_key_t{}
	instanceKey.type_ = C.CString("instance-type")
	instanceKey.url = C.CString("instance-url")
	SubmitStopSnapshot(checkId, &instanceKey)

	expectedTopology := mockTransactionalBatcher.CollectedTopology.Flush()
	instance := topology.Instance{Type: "instance-type", URL: "instance-url"}

	assert.ObjectsAreEqualValues(expectedTopology, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
				Topology: &topology.Topology{
					StartSnapshot: false,
					StopSnapshot:  true,
					Instance:      instance,
					Components:    []topology.Component{},
					Relations:     []topology.Relation{},
				},
			},
		}))
}

func testDeleteTopologyElement(t *testing.T) {
	mockTransactionalBatcher := transactionbatcher.NewMockTransactionalBatcher()

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

	expectedTopology := mockTransactionalBatcher.CollectedTopology.Flush()
	instance := topology.Instance{Type: "instance-type", URL: "instance-url"}

	assert.ObjectsAreEqualValues(expectedTopology, transactionbatcher.TransactionCheckInstanceBatchStates(
		map[check.ID]transactionbatcher.TransactionCheckInstanceBatchState{
			"check-id": {
				Topology: &topology.Topology{
					StartSnapshot: true,
					StopSnapshot:  true,
					Instance:      instance,
					Components:    []topology.Component{},
					Relations:     []topology.Relation{},
					DeleteIDs:     []string{topoElementId},
				},
			},
		}))
}
