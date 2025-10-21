package corechecks

import (
	"github.com/StackVista/stackstate-receiver-go-client/pkg/model/topology"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMakeCheckTopologyCollector(t *testing.T) {
	instance := topology.Instance{
		Type: "test",
		URL:  "url",
	}
	ptc := MakeCheckTopologyCollector(instance)
	assert.Equal(t, instance, ptc.TopologyInstance)
}
