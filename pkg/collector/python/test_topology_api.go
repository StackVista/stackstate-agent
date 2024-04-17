//go:build python && test

package python

import (
	"encoding/json"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	"github.com/DataDog/datadog-agent/pkg/collector/check/handler"
	"github.com/DataDog/datadog-agent/pkg/collector/transactional/transactionbatcher"
	"github.com/DataDog/datadog-agent/pkg/health"
	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// #include <datadog_agent_rtloader.h>
import "C"

func testComponentTopology(t *testing.T) {
	SetupTransactionalComponents()
	mockTransactionalBatcher := transactionbatcher.GetTransactionalBatcher().(*transactionbatcher.MockTransactionalBatcher)

	testCheck := &check.STSTestCheck{Name: "check-id-component-test"}
	handler.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	c := &topology.Component{
		ExternalID: "external-id",
		Type:       topology.Type{Name: "component-type"},
		Data: map[string]interface{}{
			"some": "data",
		},
	}
	data, err := json.Marshal(c)
	assert.NoError(t, err)

	checkId := C.CString(testCheck.String())
	instanceKey := C.instance_key_t{}
	instanceKey.type_ = C.CString("instance-type")
	instanceKey.url = C.CString("instance-url")

	StartTransaction(checkId)
	SubmitStartSnapshot(checkId, &instanceKey)
	SubmitComponent(
		checkId,
		&instanceKey,
		C.CString("external-id"),
		C.CString("component-type"),
		C.CString(string(data)))
	SubmitStopSnapshot(checkId, &instanceKey)

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())
	expectedTopology := transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: actualTopology.Transaction, // not asserting this specifically, it just needs to be present
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

	handler.GetCheckManager().UnsubscribeCheckHandler(testCheck.ID())
}

func testRelationTopology(t *testing.T) {
	SetupTransactionalComponents()
	mockTransactionalBatcher := transactionbatcher.GetTransactionalBatcher().(*transactionbatcher.MockTransactionalBatcher)

	testCheck := &check.STSTestCheck{Name: "check-id-relation-test"}
	handler.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

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

	checkId := C.CString(testCheck.String())
	instanceKey := C.instance_key_t{}
	instanceKey.type_ = C.CString("instance-type")
	instanceKey.url = C.CString("instance-url")

	StartTransaction(checkId)
	SubmitRelation(
		checkId,
		&instanceKey,
		C.CString("source-id"),
		C.CString("target-id"),
		C.CString("relation-type"),
		C.CString(string(data)),
	)

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())
	expectedTopology := transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: actualTopology.Transaction, // not asserting this specifically, it just needs to be present
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

	handler.GetCheckManager().UnsubscribeCheckHandler(testCheck.ID())
}

func testStartSnapshotCheck(t *testing.T) {
	SetupTransactionalComponents()
	mockTransactionalBatcher := transactionbatcher.GetTransactionalBatcher().(*transactionbatcher.MockTransactionalBatcher)

	testCheck := &check.STSTestCheck{Name: "check-id-start-snapshot"}
	handler.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	checkId := C.CString(testCheck.String())
	instanceKey := C.instance_key_t{}
	instanceKey.type_ = C.CString("instance-type")
	instanceKey.url = C.CString("instance-url")

	StartTransaction(checkId)
	SubmitStartSnapshot(checkId, &instanceKey)

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())

	assert.Equal(t, transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: actualTopology.Transaction, // not asserting this specifically, it just needs to be present
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

	handler.GetCheckManager().UnsubscribeCheckHandler(testCheck.ID())
}

func testStopSnapshotCheck(t *testing.T) {
	SetupTransactionalComponents()
	mockTransactionalBatcher := transactionbatcher.GetTransactionalBatcher().(*transactionbatcher.MockTransactionalBatcher)

	testCheck := &check.STSTestCheck{Name: "check-id-stop-snapshot"}
	handler.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	checkId := C.CString(testCheck.String())
	instanceKey := C.instance_key_t{}
	instanceKey.type_ = C.CString("instance-type")
	instanceKey.url = C.CString("instance-url")

	StartTransaction(checkId)
	SubmitStopSnapshot(checkId, &instanceKey)

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())

	expectedTopology := transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: actualTopology.Transaction, // not asserting this specifically, it just needs to be present
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

	handler.GetCheckManager().UnsubscribeCheckHandler(testCheck.ID())
}

func testDeleteTopologyElement(t *testing.T) {
	SetupTransactionalComponents()
	mockTransactionalBatcher := transactionbatcher.GetTransactionalBatcher().(*transactionbatcher.MockTransactionalBatcher)

	testCheck := &check.STSTestCheck{Name: "check-id-delete-element"}
	handler.GetCheckManager().RegisterCheckHandler(testCheck, integration.Data{}, integration.Data{})

	checkID := C.CString(testCheck.String())
	instanceKey := C.instance_key_t{}
	instanceKey.type_ = C.CString("instance-type")
	instanceKey.url = C.CString("instance-url")
	topoElementId := "topo-element-id"

	StartTransaction(checkID)
	SubmitStartSnapshot(checkID, &instanceKey)
	SubmitDelete(
		checkID,
		&instanceKey,
		C.CString(topoElementId))
	SubmitStopSnapshot(checkID, &instanceKey)

	time.Sleep(50 * time.Millisecond) // sleep a bit for everything to complete

	actualTopology, found := mockTransactionalBatcher.GetCheckState(testCheck.ID())
	assert.True(t, found, "no TransactionCheckInstanceBatchState found for check: %s", testCheck.ID())
	expectedTopology := transactionbatcher.TransactionCheckInstanceBatchState{
		Transaction: actualTopology.Transaction, // not asserting this specifically, it just needs to be present
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

	handler.GetCheckManager().UnsubscribeCheckHandler(testCheck.ID())
}
