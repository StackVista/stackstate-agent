//go:build linux
// +build linux

package modules

import (
	"fmt"
	"net/http"

	"github.com/StackVista/stackstate-agent/cmd/system-probe/api/module"
	"github.com/StackVista/stackstate-agent/cmd/system-probe/config"
	"github.com/StackVista/stackstate-agent/cmd/system-probe/utils"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/ebpf/probe"
	"github.com/StackVista/stackstate-agent/pkg/ebpf"
)

// TCPQueueLength Factory
var TCPQueueLength = module.Factory{
	Name: config.TCPQueueLengthTracerModule,
	Fn: func(cfg *config.Config) (module.Module, error) {
		t, err := probe.NewTCPQueueLengthTracer(ebpf.NewConfig())
		if err != nil {
			return nil, fmt.Errorf("unable to start the TCP queue length tracer: %w", err)
		}

		return &tcpQueueLengthModule{t}, nil
	},
}

var _ module.Module = &tcpQueueLengthModule{}

type tcpQueueLengthModule struct {
	*probe.TCPQueueLengthTracer
}

func (t *tcpQueueLengthModule) Register(httpMux *module.Router) error {
	httpMux.HandleFunc("/check/tcp_queue_length", func(w http.ResponseWriter, req *http.Request) {
		stats := t.TCPQueueLengthTracer.GetAndFlush()
		utils.WriteAsJSON(w, stats)
	})

	return nil
}

func (t *tcpQueueLengthModule) GetStats() map[string]interface{} {
	return nil
}
