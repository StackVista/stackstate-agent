package main

import (
	"context"
	"encoding/json"
	"fmt"
	porterpb "github.com/StackVista/stackstate-agent/pkg/porter/proto"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"google.golang.org/grpc"
	"log"
)

func main() {
	fmt.Println("Hello client ...")

	opts := grpc.WithInsecure()
	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatal(err)
	}
	defer cc.Close()

	porter := porterpb.NewStackPorterClient(cc)
	instance := &porterpb.StackPorterInstance{
		PorterID: "Jeremy",
		Instance: &porterpb.StackInstance{
			Type: "my-type",
			Url: "my-url",
		},
	}

	resp, err := porter.KickOffSnapshot(context.Background(), instance)
	if err != nil {
		log.Fatalf("Receive response => %s\n", err)
	}
	fmt.Printf("Receive response => %s - %s\n", resp.Status, resp.Message)

	cData, err := json.Marshal(topology.Data{"Name": "My type Name"})
	req := &porterpb.PushStackComponentRequest{
		PorterID: instance.PorterID,
		Instance: &porterpb.StackInstance{Type: instance.Instance.Type, Url: instance.Instance.Url},
		Component:            &porterpb.StackComponent{
			ExternalID:           "my-external-id",
			Type:                 &porterpb.StackType{Name: "my-type-name"},
			Data: cData,
		},
	}
	resp, err = porter.PushComponent(context.Background(), req)
	if err != nil {
		log.Fatalf("Receive response => %s\n", err)
	}
	fmt.Printf("Receive response => %s - %s\n", resp.Status, resp.Message)

	rData, err := json.Marshal(topology.Data{"SourceName": "Name1", "TargetName": "Name2"})
	rreq := &porterpb.PushStackRelationRequest{
		PorterID: instance.PorterID,
		Instance: &porterpb.StackInstance{Type: instance.Instance.Type, Url: instance.Instance.Url},
		Relation:            &porterpb.StackRelation{
			ExternalID:           "id1->id2",
			SourceID: "id1",
			TargetID: "id2",
			Type:                 &porterpb.StackType{Name: "my-relation-type-name"},
			Data: rData,
		},
	}
	resp, err = porter.PushRelation(context.Background(), rreq)
	if err != nil {
		log.Fatalf("Receive response => %s\n", err)
	}
	fmt.Printf("Receive response => %s - %s\n", resp.Status, resp.Message)

	resp, err = porter.EndSnapshot(context.Background(), instance)
	if err != nil {
		log.Fatalf("Receive response => %s\n", err)
	}
	fmt.Printf("Receive response => %s - %s\n", resp.Status, resp.Message)

	status := &porterpb.StackPorterStatus{
		Status: porterpb.PorterStatus_success,
		Message: "Completed Porter",
	}
	completePorter := &porterpb.StackPorterComplete{Instance: instance, Status: status}
	resp, err = porter.CompletePorter(context.Background(), completePorter)
	if err != nil {
		log.Fatalf("Receive response => %s\n", err)
	}
	fmt.Printf("Receive response => %s - %s\n", resp.Status, resp.Message)
}
