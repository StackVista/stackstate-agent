// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package windowsevent

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/StackVista/stackstate-agent/pkg/logs/config"
)

func TestShouldSanitizeConfig(t *testing.T) {
	launcher := NewLauncher(config.NewLogSources(), nil)
	assert.Equal(t, "*", launcher.sanitizedConfig(&config.LogsConfig{ChannelPath: "System", Query: ""}).Query)
}
