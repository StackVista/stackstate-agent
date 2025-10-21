//go:build kubeapiserver

package topologycollectors

import (
	"fmt"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	"testing"

	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var creationTime v1.Time
var replicas int32
var pathType coreV1.HostPathType
var gcePersistentDisk coreV1.GCEPersistentDiskVolumeSource
var awsElasticBlockStore coreV1.AWSElasticBlockStoreVolumeSource
var hostPath coreV1.HostPathVolumeSource

func RunCorrelatorTest(t *testing.T, correlator ClusterTopologyCorrelator, expectedCorrelatorName string) {
	actualCorrelatorName := correlator.GetName()
	assert.Equal(t, expectedCorrelatorName, actualCorrelatorName)

	// Trigger Correlator Function
	go func() {
		log.Debugf("Starting cluster topology correlator: %s\n", correlator.GetName())
		err := correlator.CorrelateFunction()
		// assert no error occurred
		assert.Nil(t, err)
		// mark this correlator as complete
		log.Debugf("Finished cluster topology correlator: %s\n", correlator.GetName())
	}()
}

func simpleRelation(sourceID string, targetID string, typ string) *topology.Relation {
	return &topology.Relation{
		ExternalID: fmt.Sprintf("%s->%s", sourceID, targetID),
		SourceID:   sourceID,
		TargetID:   targetID,
		Type:       topology.Type{Name: typ},
		Data:       map[string]interface{}{},
	}
}

func simpleRelationWithData(sourceID string, targetID string, typ string, data map[string]interface{}) *topology.Relation {
	return &topology.Relation{
		ExternalID: fmt.Sprintf("%s->%s", sourceID, targetID),
		SourceID:   sourceID,
		TargetID:   targetID,
		Type:       topology.Type{Name: typ},
		Data:       data,
	}
}
