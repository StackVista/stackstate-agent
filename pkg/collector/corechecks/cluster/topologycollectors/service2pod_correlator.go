//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/StackVista/stackstate-agent/pkg/topology"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// Service2PodCorrelator
type Service2PodCorrelator struct {
	ComponentChan       chan<- *topology.Component
	RelationChan        chan<- *topology.Relation
	PodCorrChan         <-chan *PodEndpointCorrelation
	EndpointCorrChannel <-chan *ServiceEndpointCorrelation
	ClusterTopologyCorrelator
}

// PodEndpointCorrelation an endpoint served by a pod
type PodEndpointCorrelation struct {
	Endpoint     string
	PodNamespace string
	PodName      string
}

// ServiceEndpointCorrelation an underlying endpoint for a service
type ServiceEndpointCorrelation struct {
	ServiceExternalID string
	Endpoint          EndpointID
}

// NewService2PodCorrelator creates correlator creates relation from service to pod
// in case where service points to kubernetes host and pod is exposed within the host's network
func NewService2PodCorrelator(
	componentChannel chan<- *topology.Component,
	relationChannel chan<- *topology.Relation,
	podCorrChannel chan *PodEndpointCorrelation,
	serviceCorrChannel chan *ServiceEndpointCorrelation,
	clusterTopologyCorrelator ClusterTopologyCorrelator,
) ClusterTopologyCorrelator {
	return &Service2PodCorrelator{
		ComponentChan:             componentChannel,
		RelationChan:              relationChannel,
		PodCorrChan:               podCorrChannel,
		EndpointCorrChannel:       serviceCorrChannel,
		ClusterTopologyCorrelator: clusterTopologyCorrelator,
	}
}

// GetName returns the name of the Collector
func (Service2PodCorrelator) GetName() string {
	return "Pod to Service Correlator"
}

type podID struct {
	Namespace string
	Name      string
}

// CollectorFunction collects all endpoints exposed by pods within host network
// then it collects all unmatched endpoints,
// and then it creates corresponding relations
func (crl *Service2PodCorrelator) CorrelateFunction() error {

	// making a map from a host's endpoint (x.x.x.x:yyyy) to a pod that is serving it
	podsExposedFromHost := map[string]podID{}
	for podCorr := range crl.PodCorrChan {
		podsExposedFromHost[podCorr.Endpoint] = podID{
			Namespace: podCorr.PodNamespace,
			Name:      podCorr.PodName,
		}
	}

	// now for every endpoint we are trying to find a pod serving it
	for svcCorr := range crl.EndpointCorrChannel {
		serviceID := svcCorr.ServiceExternalID
		endpointID := svcCorr.Endpoint.URL

		if pod, ok := podsExposedFromHost[endpointID]; ok {
			podID := crl.buildPodExternalID(pod.Namespace, pod.Name)
			relation := crl.CreateRelation(serviceID, podID, "exposes")
			log.Tracef("Correlated StackState service -> pod relation %s->%s", relation.SourceID, relation.TargetID)
			crl.RelationChan <- relation
		}
	}

	return nil
}
