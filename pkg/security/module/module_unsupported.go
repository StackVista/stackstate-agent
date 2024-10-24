// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build !linux
// +build !linux

package module

import (
	"github.com/StackVista/stackstate-agent/cmd/system-probe/api/module"
	"github.com/StackVista/stackstate-agent/pkg/ebpf"
	"github.com/StackVista/stackstate-agent/pkg/security/config"
)

// NewModule instantiates a runtime security system-probe module
func NewModule(cfg *config.Config) (module.Module, error) {
	return nil, ebpf.ErrNotImplemented
}
