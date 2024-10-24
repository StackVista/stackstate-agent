// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package config

import (
	"path/filepath"

	"github.com/StackVista/stackstate-agent/pkg/util/winutil"
)

var (
	defaultConfdPath            = "c:\\programdata\\datadog\\conf.d"
	defaultAdditionalChecksPath = "c:\\programdata\\datadog\\checks.d"
	defaultRunPath              = "c:\\programdata\\datadog\\run"
	defaultSyslogURI            = ""
	defaultGuiPort              = 5002
	// defaultSecurityAgentLogFile points to the log file that will be used by the security-agent if not configured
	defaultSecurityAgentLogFile = "c:\\programdata\\datadog\\logs\\security-agent.log"
	// defaultSystemProbeAddress is the default address to be used for connecting to the system probe
	defaultSystemProbeAddress     = "localhost:3333"
	defaultSystemProbeLogFilePath = "c:\\programdata\\datadog\\logs\\system-probe.log"
)

// ServiceName is the name that'll be used to register the Agent
const ServiceName = "DatadogAgent"

func osinit() {
	pd, err := winutil.GetProgramDataDir()
	if err == nil {
		defaultConfdPath = filepath.Join(pd, "conf.d")
		defaultAdditionalChecksPath = filepath.Join(pd, "checks.d")
		defaultRunPath = filepath.Join(pd, "run")
		defaultSecurityAgentLogFile = filepath.Join(pd, "logs", "security-agent.log")
		defaultSystemProbeLogFilePath = filepath.Join(pd, "logs", "system-probe.log")
	} else {
		winutil.LogEventViewer(ServiceName, 0x8000000F, defaultConfdPath)
	}
}

// NewAssetFs  Should never be called on non-android
func setAssetFs(config Config) {}
