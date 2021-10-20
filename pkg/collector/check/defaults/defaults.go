// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package defaults

import (
	"time"
)

const (
	// DefaultCheckInterval is the interval in seconds the scheduler should apply
	// when no value was provided in Check configuration.
  	// Be sure to also update stackstate_checks_base/stackstate_checks/base/checks/base.py
  	// in stackstate-agent-integrations when the default is changed
	DefaultCheckInterval time.Duration = 40 * time.Second
)
