// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package app

import (
	"github.com/StackVista/stackstate-agent/cmd/agent/common/commands"
)

func init() {
	AgentCmd.AddCommand(commands.Health(loggerName, &confFilePath, &flagNoColor))
}
