package porter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	porterpb "github.com/StackVista/stackstate-agent/pkg/porter/proto"
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"google.golang.org/grpc"
	"log"
	"net"
)

// StackPorterServer wraps the porter gRPC server
type StackPorterServer struct {
	Server *grpc.Server
}

// Start starts doing the gRPC server and is ready to receive data from porters
func (s *StackPorterServer) Start(mainCtx context.Context) {
	defer s.Stop()

	address := "0.0.0.0:50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	fmt.Printf("Server is listening on %v ...", address)

	s.Server = grpc.NewServer()

	porterpb.RegisterStackPorterServer(s.Server, &StackPorterServer{})

	go func() {
		err = s.Server.Serve(lis)
		if err != nil {
			log.Fatalf("Error %v", err)
		}
	}()

	<-mainCtx.Done()
}

// Stop stops the porter gRPC server
func (s *StackPorterServer) Stop() {
	s.Server.GracefulStop()
}

// KickOffSnapshot pushes a SubmitStartSnapshot to the batcher instance
func (s *StackPorterServer) KickOffSnapshot(ctx context.Context, req *porterpb.StackPorterInstance) (*porterpb.StackPorterStatus, error) {
	batcher.GetBatcher().SubmitStartSnapshot(check.ID(req.PorterID), s.CreateInstance(req.Instance))

	return &porterpb.StackPorterStatus{
		Status:  porterpb.PorterStatus_success,
		Message: "Submitted start snapshot",
	}, nil
}

// EndSnapshot pushes a SubmitStopSnapshot to the batcher instance
func (s *StackPorterServer) EndSnapshot(ctx context.Context, req *porterpb.StackPorterInstance) (*porterpb.StackPorterStatus, error) {
	batcher.GetBatcher().SubmitStopSnapshot(check.ID(req.PorterID), s.CreateInstance(req.Instance))

	return &porterpb.StackPorterStatus{
		Status:  porterpb.PorterStatus_success,
		Message: "Submitted stop snapshot",
	}, nil
}

// CompletePorter pushes a SubmitComplete to the batcher instance
func (s *StackPorterServer) CompletePorter(ctx context.Context, req *porterpb.StackPorterComplete) (*porterpb.StackPorterStatus, error) {
	batcher.GetBatcher().SubmitComplete(check.ID(req.PorterID))

	return &porterpb.StackPorterStatus{
		Status:  porterpb.PorterStatus_success,
		Message: "Submitted complete for porter",
	}, nil
}

// PushComponent pushes a component to the batcher instance
func (s *StackPorterServer) PushComponent(ctx context.Context, req *porterpb.PushStackComponentRequest) (*porterpb.StackPorterStatus, error) {
	batcher.GetBatcher().SubmitComponent(check.ID(req.PorterID), s.CreateInstance(req.Instance), s.CreateComponent(req.Component))

	return &porterpb.StackPorterStatus{
		Status:  porterpb.PorterStatus_success,
		Message: fmt.Sprintf("Submitted component for porter %s", req.PorterID),
	}, nil
}

// PushRelation pushes a relation to the batcher instance
func (s *StackPorterServer) PushRelation(ctx context.Context, req *porterpb.PushStackRelationRequest) (*porterpb.StackPorterStatus, error) {
	batcher.GetBatcher().SubmitRelation(check.ID(req.PorterID), s.CreateInstance(req.Instance), s.CreateRelation(req.Relation))

	return &porterpb.StackPorterStatus{
		Status:  porterpb.PorterStatus_success,
		Message: fmt.Sprintf("Submitted relation for porter %s", req.PorterID),
	}, nil
}

// CreateInstance return a topology instance
func (*StackPorterServer) CreateInstance(instance *porterpb.StackInstance) topology.Instance {
	return topology.Instance{
		Type: instance.Type,
		URL:  instance.Url,
	}
}

// CreateComponent returns a topology component
func (*StackPorterServer) CreateComponent(component *porterpb.StackComponent) topology.Component {
	// get the data
	var data topology.Data
	err := json.Unmarshal(component.Data, &data)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	return topology.Component{
		ExternalID: component.ExternalID,
		Type:       topology.Type{Name: component.Type.Name},
		Data:       data,
	}
}

// CreateRelation returns a topology relation
func (*StackPorterServer) CreateRelation(relation *porterpb.StackRelation) topology.Relation {
	// get the data
	var data topology.Data
	err := json.Unmarshal(relation.Data, &data)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	return topology.Relation{
		ExternalID: relation.ExternalID,
		SourceID:   relation.SourceID,
		TargetID:   relation.TargetID,
		Type:       topology.Type{Name: relation.Type.Name},
		Data:       data,
	}
}
