//go:build !linux && !windows
// +build !linux,!windows

package modules

import "github.com/StackVista/stackstate-agent/cmd/system-probe/api/module"

// All System Probe modules should register their factories here
var All = []module.Factory{}
