package selfcheck

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

func (sc *selfCheckTopology) ID() check.ID {
	return check.ID("agent-checks-check")
}

func (sc *selfCheckTopology) Instance() topology.Instance {
	return topology.Instance{
		Type: "agent",
		URL:  "integrations",
	}
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

func (sc *selfCheckTopology) CheckID(id check.ID) string {
	return fmt.Sprintf("%s:%s", sc.Hostname, id)
}

func (sc *selfCheckTopology) CheckComponent(ch check.Check) *topology.Component {
	return &topology.Component{
		ExternalID: sc.CheckID(ch.ID()),
		Type: topology.Type{
			Name: "agent-integration",
		},
		Data: topology.Data{
			"name":               fmt.Sprintf("%s", ch.ID()),
			"layer":              "Checks",
			"interval":           ch.Interval(),
			"configSource":       ch.ConfigSource(),
			"config":             ch.GetConfiguration(),
			"version":            ch.Version(),
			"isTelemetryEnabled": ch.IsTelemetryEnabled(),
			"tags":               []string{"agent-integration", "self-observability"},
		},
	}
}

func (sc *selfCheckTopology) SyncComponent(instance topology.Instance) *topology.Component {
	return &topology.Component{
		ExternalID: instance.GoString(),
		Type: topology.Type{
			Name: "topology-sync",
		},
		Data: topology.Data{
			"name":  fmt.Sprintf("%s %s", instance.Type, instance.URL),
			"type":  instance.Type,
			"layer": "Synchronization",
			"url":   instance.URL,
			"tags":  []string{"agent-integration", "self-observability"},
		},
	}
}

func (sc *selfCheckTopology) SyncToCheckRelation(instance topology.Instance, checkID check.ID) *topology.Relation {
	return &topology.Relation{
		ExternalID: fmt.Sprintf("%s>%s", sc.AgentID(), checkID),
		SourceID:   instance.GoString(),
		TargetID:   sc.CheckID(checkID),
		Type: topology.Type{
			Name: "populated_with",
		},
		Data: topology.Data{},
	}
}

func (sc *selfCheckTopology) CheckToAgentRelation(ch check.Check) *topology.Relation {
	return &topology.Relation{
		ExternalID: fmt.Sprintf("%s>%s", sc.AgentID(), sc.CheckID(ch.ID())),
		SourceID:   sc.AgentID(),
		TargetID:   sc.CheckID(ch.ID()),
		Type: topology.Type{
			Name: "runs",
		},
		Data: topology.Data{},
	}
}
