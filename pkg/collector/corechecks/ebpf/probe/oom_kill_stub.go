//go:build !linux_bpf
// +build !linux_bpf

package probe

import (
	"github.com/StackVista/stackstate-agent/pkg/ebpf"
)

// OOMKillProbe is not implemented on non-linux systems
type OOMKillProbe struct{}

// NewOOMKillProbe is not implemented on non-linux systems
func NewOOMKillProbe(cfg *ebpf.Config) (*OOMKillProbe, error) {
	return nil, ebpf.ErrNotImplemented
}

// Close is not implemented on non-linux systems
func (t *OOMKillProbe) Close() {}

// Get is not implemented on non-linux systems
func (t *OOMKillProbe) Get() []OOMKillStats {
	return nil
}

// GetAndFlush is not implemented on non-linux systems
func (t *OOMKillProbe) GetAndFlush() []OOMKillStats {
	return nil
}
