// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package util

import (
	"fmt"

	"github.com/StackVista/stackstate-agent/pkg/config"
)

// SetupCoreDump enables core dumps and sets the core dump size limit based on configuration
func SetupCoreDump() error {
	if config.Datadog.GetBool("go_core_dump") {
		return fmt.Errorf("Not supported on Windows")
	}
	return nil
}
