package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/porter"
	"github.com/fatih/color"
)

func main() {
	mainCtx, mainCtxCancel := context.WithCancel(context.Background())
	defer mainCtxCancel() // Calling cancel twice is safe

	// [sts] init the batcher without the real serializer
	batcher.InitBatcher(&printingAgentV1Serializer{}, "my-hostname", "agent", config.GetMaxCapacity())

	grpcServer := &porter.StackPorterServer{}
	grpcServer.Start(mainCtx)
}

type printingAgentV1Serializer struct{}

func (printingAgentV1Serializer) SendJSONToV1Intake(data interface{}) error {
	fmt.Fprintln(color.Output, fmt.Sprintf("=== %s ===", color.BlueString("Topology")))
	j, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(j))
	return nil
}
