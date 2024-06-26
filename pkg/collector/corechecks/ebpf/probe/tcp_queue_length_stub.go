//go:build !linux_bpf
// +build !linux_bpf

package probe

import (
	"github.com/StackVista/stackstate-agent/pkg/ebpf"
)

// TCPQueueLengthTracer is not implemented on non-linux systems
type TCPQueueLengthTracer struct{}

// NewTCPQueueLengthTracer is not implemented on non-linux systems
func NewTCPQueueLengthTracer(cfg *ebpf.Config) (*TCPQueueLengthTracer, error) {
	return nil, ebpf.ErrNotImplemented
}

// Close is not implemented on non-linux systems
func (t *TCPQueueLengthTracer) Close() {}

// Get is not implemented on non-linux systems
func (t *TCPQueueLengthTracer) Get() []TCPQueueLengthStats {
	return nil
}

// GetAndFlush is not implemented on non-linux systems
func (t *TCPQueueLengthTracer) GetAndFlush() []TCPQueueLengthStats {
	return nil
}
