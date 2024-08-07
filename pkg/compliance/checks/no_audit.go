// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build !linux
// +build !linux

package checks

import (
	"errors"

	"github.com/StackVista/stackstate-agent/pkg/compliance/checks/env"
)

func newAuditClient() (env.AuditClient, error) {
	return nil, errors.New("audit client requires linux build flag")
}
