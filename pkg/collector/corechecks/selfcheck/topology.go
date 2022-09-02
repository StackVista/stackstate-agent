package selfcheck

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	collectorutils "github.com/StackVista/stackstate-agent/pkg/collector/util"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util"
	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/clustername"
	"github.com/StackVista/stackstate-agent/pkg/version"
	"os"
)

const SelfCheckID check.ID = "agent-checks-check"

type selfCheckTopology struct {
	Hostname string
}

func NewSelfCheckTopology() *selfCheckTopology {
	hostname, _ := util.GetHostname()
	return &selfCheckTopology{Hostname: hostname}
}

func (sc *selfCheckTopology) ID() check.ID {
	return SelfCheckID
}

func (sc *selfCheckTopology) Instance() topology.Instance {
	return topology.Instance{
		Type: "agent",
		URL:  "integrations",
	}
}

func (sc *selfCheckTopology) AgentID() string {
	return fmt.Sprintf("urn:process:/%s:%d:%d", sc.Hostname, os.Getpid(), getCurrentProcessCreateTime())
}

func getCurrentProcessCreateTime() int64 {
	pid := os.Getpid()
	createTime, _ := collectorutils.GetProcessCreateTime(int32(pid))
	return createTime
}

func (sc *selfCheckTopology) AgentComponent() *topology.Component {
	name := fmt.Sprintf("StackState Agent:%s", sc.Hostname)
	return &topology.Component{
		ExternalID: sc.AgentID(),
		Type: topology.Type{
			Name: "stackstate-agent",
		},
		Data: topology.Data{
			"name":     name,
			"version":  version.AgentVersion,
			"hostname": sc.Hostname,
			"cluster":  clustername.GetClusterName(),
			"identifiers": []string{
				sc.AgentID(),
			},
			"tags": []string{
				"stackstate-agent", "agent-integration", "self-observability",
				fmt.Sprintf("hostname:%s", sc.Hostname),
			},
		},
	}
}

func (sc *selfCheckTopology) CheckID(id check.ID) string {
	return fmt.Sprintf("urn:agent-integration:/%s:%s", sc.Hostname, id)
}

func (sc *selfCheckTopology) CheckComponent(ch check.Check) *topology.Component {
	return &topology.Component{
		ExternalID: sc.CheckID(ch.ID()),
		Type: topology.Type{
			Name: "agent-integration",
		},
		Data: topology.Data{
			"name":               fmt.Sprintf("%s", ch.ID()),
			"hostname":           sc.Hostname,
			"interval":           ch.Interval(),
			"config":             ch.GetConfiguration(),
			"version":            ch.Version(),
			"isTelemetryEnabled": ch.IsTelemetryEnabled(),
			"identifiers": []string{
				sc.CheckID(ch.ID()),
			},
			"tags": []string{"agent-integration", "self-observability"},
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
		ExternalID: fmt.Sprintf("%s>%s", sc.CheckID(ch.ID()), sc.AgentID()),
		SourceID:   sc.CheckID(ch.ID()),
		TargetID:   sc.AgentID(),
		Type: topology.Type{
			Name: "run_with",
		},
		Data: topology.Data{},
	}
}
