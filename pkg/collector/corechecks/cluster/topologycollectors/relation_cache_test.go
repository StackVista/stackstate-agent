//go:build kubeapiserver

package topologycollectors

import (
	"fmt"
	"sync"
	"testing"

	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	"github.com/stretchr/testify/assert"
)

// Collectors and correlators submit relations from different goroutines; none
// may be lost while their endpoints are still unknown.
func TestRelationCacheKeepsConcurrentlySubmittedRelations(t *testing.T) {
	const submitters, perSubmitter = 8, 200
	componentChan := make(chan *topology.Component)
	relationChan := make(chan *topology.Relation)
	common := NewClusterTopologyCommon(topology.Instance{Type: "kubernetes", URL: "test"}, Kubernetes, nil,
		componentChan, relationChan, nil)

	var submitted sync.WaitGroup
	for s := 0; s < submitters; s++ {
		submitted.Add(1)
		go func(s int) {
			defer submitted.Done()
			for i := 0; i < perSubmitter; i++ {
				common.SubmitRelation(common.CreateRelation(fmt.Sprintf("source-%d-%d", s, i), "target", "uses"))
			}
		}(s)
	}
	submitted.Wait()

	go func() {
		common.SubmitComponent(&topology.Component{ExternalID: "target"})
		for s := 0; s < submitters; s++ {
			for i := 0; i < perSubmitter; i++ {
				common.SubmitComponent(&topology.Component{ExternalID: fmt.Sprintf("source-%d-%d", s, i)})
			}
		}
		common.CorrelateRelations()
		close(relationChan)
	}()

	received := 0
	for {
		select {
		case <-componentChan:
		case _, ok := <-relationChan:
			if !ok {
				assert.Equal(t, submitters*perSubmitter, received)
				return
			}
			received++
		}
	}
}
