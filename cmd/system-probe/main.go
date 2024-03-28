//go:build linux
// +build linux

package main

import (
	"os"

	"github.com/StackVista/stackstate-agent/cmd/system-probe/app"
)

func main() {
	setDefaultCommandIfNonePresent()
	checkForDeprecatedFlags()
	if err := app.SysprobeCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
