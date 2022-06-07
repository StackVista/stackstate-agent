package modules

import (
	"net/http"

	"github.com/StackVista/stackstate-agent/cmd/system-probe/api"
	"github.com/StackVista/stackstate-agent/cmd/system-probe/utils"
	"github.com/StackVista/stackstate-agent/pkg/ebpf"
	"github.com/StackVista/stackstate-agent/pkg/process/config"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/pkg/errors"
)

// TCPQueueLength Factory
var TCPQueueLength = api.Factory{
	Name: "tcp_queue_length_tracer",
	Fn: func(cfg *config.AgentConfig) (api.Module, error) {
		if !cfg.CheckIsEnabled("TCP queue length") {
			log.Infof("TCP queue length tracer disabled")
			return nil, api.ErrNotEnabled
		}

		t, err := ebpf.NewTCPQueueLengthTracer(config.SysProbeConfigFromConfig(cfg))
		if err != nil {
			return nil, errors.Wrapf(err, "unable to start the TCP queue length tracer")
		}

		return &tcpQueueLengthModule{t}, nil
	},
}

var _ api.Module = &tcpQueueLengthModule{}

type tcpQueueLengthModule struct {
	*ebpf.TCPQueueLengthTracer
}

func (t *tcpQueueLengthModule) Register(httpMux *http.ServeMux) error {
	httpMux.HandleFunc("/check/tcp_queue_length", func(w http.ResponseWriter, req *http.Request) {
		stats := t.TCPQueueLengthTracer.GetAndFlush()
		utils.WriteAsJSON(w, stats)
	})

	return nil
}

func (t *tcpQueueLengthModule) GetStats() map[string]interface{} {
	return nil
}
