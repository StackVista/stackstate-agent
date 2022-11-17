//go:build linux_bpf
// +build linux_bpf

package kprobe

import (
	"testing"

	"github.com/StackVista/stackstate-agent/pkg/ebpf/bytecode/runtime"
	"github.com/StackVista/stackstate-agent/pkg/network/config"
	"github.com/stretchr/testify/require"
)

func TestTracerCompile(t *testing.T) {
	cfg := config.New()
	cfg.BPFDebug = true
	cflags := getCFlags(cfg)
	_, err := runtime.Tracer.Compile(&cfg.Config, cflags)
	require.NoError(t, err)
}
