// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build !windows

package tags

import (
	"bytes"

	"golang.org/x/sys/unix"
)

// ResolveRuntimeArch determines the architecture of the lambda at runtime
func ResolveRuntimeArch() string {
	var uname unix.Utsname
	if err := unix.Uname(&uname); err != nil {
		return "amd64"
	}

	switch string(uname.Machine[:bytes.IndexByte(uname.Machine[:], 0)]) {
	case "aarch64":
		return "arm64"
	default:
		return "x86_64"
	}
}
