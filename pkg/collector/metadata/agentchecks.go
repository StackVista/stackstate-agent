// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package metadata

import (
	"context"
	"fmt"

	"github.com/StackVista/stackstate-agent/pkg/collector/metadata/agentchecks"
	md "github.com/StackVista/stackstate-agent/pkg/metadata"
	"github.com/StackVista/stackstate-agent/pkg/serializer"
)

// AgentChecksCollector fills and sends the old metadata payload used in the
// Agent v5 for agent check status
type AgentChecksCollector struct{}

// Send collects the data needed and submits the payload
func (hp *AgentChecksCollector) Send(ctx context.Context, s *serializer.Serializer) error {
	payload := agentchecks.GetPayload(ctx)
	if err := s.SendAgentchecksMetadata(payload); err != nil {
		return fmt.Errorf("unable to submit agentchecks metadata payload, %s", err)
	}
	return nil
}

func init() {
	md.RegisterCollector("agent_checks", new(AgentChecksCollector))
}
