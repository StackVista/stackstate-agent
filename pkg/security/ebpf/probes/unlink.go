// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build linux

package probes

import manager "github.com/DataDog/ebpf-manager"

// unlinkProbes holds the list of probes used to track unlink events
var unlinkProbes = []*manager.Probe{
	{
		ProbeIdentificationPair: manager.ProbeIdentificationPair{
			UID:          SecurityAgentUID,
			EBPFSection:  "kprobe/vfs_unlink",
			EBPFFuncName: "kprobe_vfs_unlink",
		},
	},
}

func getUnlinkProbes() []*manager.Probe {
	unlinkProbes = append(unlinkProbes, ExpandSyscallProbes(&manager.Probe{
		ProbeIdentificationPair: manager.ProbeIdentificationPair{
			UID: SecurityAgentUID,
		},
		SyscallFuncName: "unlink",
	}, EntryAndExit)...)
	unlinkProbes = append(unlinkProbes, ExpandSyscallProbes(&manager.Probe{
		ProbeIdentificationPair: manager.ProbeIdentificationPair{
			UID: SecurityAgentUID,
		},
		SyscallFuncName: "unlinkat",
	}, EntryAndExit)...)
	return unlinkProbes
}
