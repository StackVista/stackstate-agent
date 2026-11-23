//go:build kubeapiserver

package topologycollectors

import (
	"fmt"
	"sync"
	"testing"

	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnapshotKeepsFirstComponentAndRelation(t *testing.T) {
	for _, cacheRelations := range []bool{true, false} {
		t.Run(fmt.Sprintf("cacheRelations=%t", cacheRelations), func(t *testing.T) {
			components := make(chan *topology.Component, 10)
			relations := make(chan *topology.Relation, 10)
			common := NewClusterTopologyCommon(topology.Instance{Type: "kubernetes", URL: "test"}, Kubernetes, nil,
				components, relations, nil)
			common.SetUseRelationCache(cacheRelations)

			// A container can mount two files from the same ConfigMap. Both
			// observations describe one relation, even when endpoints arrive later.
			firstRelation := common.CreateRelationData("container", "configmap", "uses", map[string]interface{}{"mount_path": "/first"})
			common.SubmitRelation(firstRelation)
			firstComponent := &topology.Component{ExternalID: "container", Data: map[string]interface{}{"name": "first"}}
			common.SubmitComponent(firstComponent)
			common.SubmitComponent(&topology.Component{ExternalID: "configmap"})
			common.SubmitComponent(&topology.Component{ExternalID: "container", Data: map[string]interface{}{"name": "second"}})
			common.SubmitRelation(common.CreateRelationData("container", "configmap", "uses", map[string]interface{}{"mount_path": "/second"}))
			common.CorrelateRelations()
			common.CorrelateRelations()

			require.Len(t, components, 2)
			assert.Same(t, firstComponent, <-components)
			require.Len(t, relations, 1)
			assert.Same(t, firstRelation, <-relations)
		})
	}
}

func TestSnapshotDeduplicationIsConcurrentAndScopedToOneSnapshot(t *testing.T) {
	const submitters, identities = 8, 50
	components := make(chan *topology.Component, submitters*identities*2)
	relations := make(chan *topology.Relation, submitters*identities*2)
	for snapshot := 0; snapshot < 2; snapshot++ {
		common := NewClusterTopologyCommon(topology.Instance{Type: "kubernetes", URL: "test"}, Kubernetes, nil,
			components, relations, nil)
		var submitted sync.WaitGroup
		for s := 0; s < submitters; s++ {
			submitted.Add(1)
			go func() {
				defer submitted.Done()
				for i := 0; i < identities; i++ {
					id := fmt.Sprintf("component-%d", i)
					common.SubmitRelation(common.CreateRelation(id, "component-0", "uses"))
					common.SubmitComponent(&topology.Component{ExternalID: id})
				}
			}()
		}
		submitted.Wait()
		common.CorrelateRelations()
		require.Len(t, components, identities)
		require.Len(t, relations, identities)
		seenComponents, seenRelations := map[string]bool{}, map[string]bool{}
		for i := 0; i < identities; i++ {
			component, relation := <-components, <-relations
			assert.False(t, seenComponents[component.ExternalID])
			assert.False(t, seenRelations[relation.ExternalID])
			seenComponents[component.ExternalID], seenRelations[relation.ExternalID] = true, true
		}
	}
}

func TestSnapshotStillDropsRelationsWithMissingEndpoints(t *testing.T) {
	components := make(chan *topology.Component, 2)
	relations := make(chan *topology.Relation, 2)
	common := NewClusterTopologyCommon(topology.Instance{Type: "kubernetes", URL: "test"}, Kubernetes, nil,
		components, relations, nil)
	common.SubmitComponent(&topology.Component{ExternalID: "source"})
	common.SubmitRelation(common.CreateRelation("source", "missing", "uses"))
	common.SubmitRelation(common.CreateRelation("missing", "source", "uses"))
	common.CorrelateRelations()
	assert.Empty(t, relations)
}
