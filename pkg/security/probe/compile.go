//go:build linux_bpf && !ebpf_bindata
// +build linux_bpf,!ebpf_bindata

package probe

import (
	"github.com/StackVista/stackstate-agent/pkg/ebpf/bytecode"
	"github.com/StackVista/stackstate-agent/pkg/ebpf/bytecode/runtime"
	"github.com/StackVista/stackstate-agent/pkg/security/config"
)

// TODO change probe.c path to runtime-compilation specific version
//go:generate go run ../../ebpf/include_headers.go ../ebpf/c/prebuilt/probe.c ../../ebpf/bytecode/build/runtime/runtime-security.c ../ebpf/c ../../ebpf/c
//go:generate go run ../../ebpf/bytecode/runtime/integrity.go ../../ebpf/bytecode/build/runtime/runtime-security.c ../../ebpf/bytecode/runtime/runtime-security.go runtime

func getRuntimeCompiledProbe(config *config.Config, useSyscallWrapper bool) (bytecode.AssetReader, error) {
	var cflags []string

	if useSyscallWrapper {
		cflags = append(cflags, "-DUSE_SYSCALL_WRAPPER=1")
	} else {
		cflags = append(cflags, "-DUSE_SYSCALL_WRAPPER=0")
	}

	return runtime.RuntimeSecurity.Compile(&config.Config, cflags)
}
