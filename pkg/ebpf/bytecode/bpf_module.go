//go:build linux_bpf
// +build linux_bpf

package bytecode

import (
	"fmt"
)

// ReadBPFModule from the asset file
func ReadBPFModule(bpfDir string, debug bool) (*bpflib.Module, error) {
	file := "pkg/ebpf/c/tracer-ebpf.o"
	if debug {
		file = "pkg/ebpf/c/tracer-ebpf-debug.o"
	}

	ebpfReader, err := GetReader(bpfDir, file)
	if err != nil {
		return nil, fmt.Errorf("couldn't find asset: %s", err)
	}

	m := bpflib.NewModuleFromReader(ebpfReader)
	if m == nil {
		return nil, fmt.Errorf("BPF not supported")
	}

	ebpfReader, err := GetReader(bpfDir, file)
	if err != nil {
		return nil, fmt.Errorf("couldn't find asset: %s", err)
	}

	return ebpfReader, nil
}
