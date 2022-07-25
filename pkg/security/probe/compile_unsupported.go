//go:build (linux && !linux_bpf) || ebpf_bindata
// +build linux,!linux_bpf ebpf_bindata

package probe

import (
	"fmt"

	"github.com/StackVista/stackstate-agent/pkg/ebpf/bytecode"
	"github.com/StackVista/stackstate-agent/pkg/security/config"
)

func getRuntimeCompiledProbe(config *config.Config, useSyscallWrapper bool) (bytecode.AssetReader, error) {
	return nil, fmt.Errorf("runtime compilation unsupported")
}
