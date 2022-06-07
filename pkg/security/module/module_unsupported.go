// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

//go:build !linux_bpf
// +build !linux_bpf

package module

import (
	"github.com/StackVista/stackstate-agent/cmd/system-probe/api"
	"github.com/StackVista/stackstate-agent/pkg/ebpf"
	aconfig "github.com/StackVista/stackstate-agent/pkg/process/config"
)

// NewModule instantiates a runtime security system-probe module
func NewModule(cfg *aconfig.AgentConfig) (api.Module, error) {
	return nil, ebpf.ErrNotImplemented
}
