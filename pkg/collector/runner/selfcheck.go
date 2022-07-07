package runner

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util"
)

type selfCheckTopology struct {
	Hostname string
}

func NewSelfCheckTopology() *selfCheckTopology {
	hostname, _ := util.GetHostname()
	return &selfCheckTopology{Hostname: hostname}
}

func (sc *selfCheckTopology) AgentID() string {
	return fmt.Sprintf("urn:stackstate-agent:/%s", sc.Hostname)
}

func (sc *selfCheckTopology) AgentComponent() *topology.Component {
	name := fmt.Sprintf("StackState Agent:%s", sc.Hostname)
	return &topology.Component{
		ExternalID: sc.AgentID(),
		Type: topology.Type{
			Name: "stackstate-agent",
		},
		Data: topology.Data{
			"name": name,
		},
	}
}

func (sc *selfCheckTopology) CheckID(ch check.Check) string {
	return fmt.Sprintf("%s:%s", sc.Hostname, ch.ID())
}

func (sc *selfCheckTopology) CheckComponent(ch check.Check) *topology.Component {
	return &topology.Component{
		ExternalID: sc.CheckID(ch),
		Type: topology.Type{
			Name: "agent-integration",
		},
		Data: topology.Data{
			"name":               fmt.Sprintf("%s check on %s", ch.ID(), sc.Hostname),
			"interval":           ch.Interval(),
			"configSource":       ch.ConfigSource(),
			"config":             ch.GetConfiguration(),
			"version":            ch.Version(),
			"isTelemetryEnabled": ch.IsTelemetryEnabled(),
			"tags":               []string{"agent-integration"},
		},
	}
}

func (sc *selfCheckTopology) CheckToAgentRelation(ch check.Check) *topology.Relation {
	return &topology.Relation{
		ExternalID: fmt.Sprintf("%s>%s", sc.AgentID(), sc.CheckID(ch)),
		SourceID:   sc.AgentID(),
		TargetID:   sc.CheckID(ch),
		Type: topology.Type{
			Name: "runs",
		},
		Data: topology.Data{},
	}
}
