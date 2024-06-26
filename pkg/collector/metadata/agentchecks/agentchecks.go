// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package agentchecks

import (
	"context"
	"encoding/json"

	"github.com/StackVista/stackstate-agent/pkg/autodiscovery"
	"github.com/StackVista/stackstate-agent/pkg/collector"
	"github.com/StackVista/stackstate-agent/pkg/collector/runner/expvars"
	"github.com/StackVista/stackstate-agent/pkg/metadata/common"
	"github.com/StackVista/stackstate-agent/pkg/metadata/externalhost"
	"github.com/StackVista/stackstate-agent/pkg/metadata/host"
	"github.com/StackVista/stackstate-agent/pkg/status"
	"github.com/StackVista/stackstate-agent/pkg/util"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// GetPayload builds a payload of all the agentchecks metadata
func GetPayload(ctx context.Context) *Payload {
	agentChecksPayload := ACPayload{}
	hostnameData, _ := util.GetHostnameData(ctx)
	hostname := hostnameData.Hostname
	checkStats := expvars.GetCheckStats()
	jmxStartupError := status.GetJMXStartupError()

	for _, stats := range checkStats {
		for _, s := range stats {
			var status []interface{}
			if s.LastError != "" {
				status = []interface{}{
					s.CheckName, s.CheckName, s.CheckID, "ERROR", s.LastError, "",
				}
			} else if len(s.LastWarnings) != 0 {
				status = []interface{}{
					s.CheckName, s.CheckName, s.CheckID, "WARNING", s.LastWarnings, "",
				}
			} else {
				status = []interface{}{
					s.CheckName, s.CheckName, s.CheckID, "OK", "", "",
				}
			}
			if status != nil {
				agentChecksPayload.AgentChecks = append(agentChecksPayload.AgentChecks, status)
			}
		}
	}

	loaderErrors := collector.GetLoaderErrors()

	for check, errs := range loaderErrors {
		jsonErrs, err := json.Marshal(errs)
		if err != nil {
			log.Warnf("Error formatting loader error from check %s: %v", check, err)
		}
		status := []interface{}{
			check, check, "initialization", "ERROR", string(jsonErrs),
		}
		agentChecksPayload.AgentChecks = append(agentChecksPayload.AgentChecks, status)
	}

	configErrors := autodiscovery.GetConfigErrors()

	for check, e := range configErrors {
		status := []interface{}{
			check, check, "initialization", "ERROR", e,
		}
		agentChecksPayload.AgentChecks = append(agentChecksPayload.AgentChecks, status)
	}

	if jmxStartupError.LastError != "" {
		status := []interface{}{
			"jmx", "jmx", "initialization", "ERROR", jmxStartupError.LastError,
		}
		agentChecksPayload.AgentChecks = append(agentChecksPayload.AgentChecks, status)
	}

	// Grab the non agent checks information
	metaPayload := host.GetMeta(ctx, hostnameData)
	metaPayload.Hostname = hostname
	cp := common.GetPayload(hostname)
	ehp := externalhost.GetPayload()
	payload := &Payload{
		CommonPayload{*cp},
		MetaPayload{*metaPayload},
		agentChecksPayload,
		ExternalHostPayload{*ehp},
	}

	return payload
}
