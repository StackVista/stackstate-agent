//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// Service2PodCorrelator
type Service2PodCorrelator struct {
	PodCorrChan         <-chan *PodEndpointCorrelation
	EndpointCorrChannel <-chan *ServiceEndpointCorrelation
	ClusterTopologyCorrelator
}

// PodEndpointCorrelation an endpoint served by a pod
type PodEndpointCorrelation struct {
	Labels       map[string]string
	PodNamespace string
	PodName      string
}

// ServiceEndpointCorrelation an underlying endpoint for a service
type ServiceEndpointCorrelation struct {
	ServiceExternalID string
	Namespace         string
	// Only labelMatchers, no labelExpressions for service selection
	LabelSelector map[string]string
}

// NewService2PodCorrelator creates correlator creates relation from service to pod
// in case where service points to kubernetes host and pod is exposed within the host's network
func NewService2PodCorrelator(
	podCorrChannel chan *PodEndpointCorrelation,
	serviceCorrChannel chan *ServiceEndpointCorrelation,
	clusterTopologyCorrelator ClusterTopologyCorrelator,
) ClusterTopologyCorrelator {
	return &Service2PodCorrelator{
		PodCorrChan:               podCorrChannel,
		EndpointCorrChannel:       serviceCorrChannel,
		ClusterTopologyCorrelator: clusterTopologyCorrelator,
	}
}

// GetName returns the name of the Collector
func (Service2PodCorrelator) GetName() string {
	return "Pod to Service Correlator"
}

// podSelectorMatchesPodLabels asserts whether a podSelector matches the podLabels. A podSelector is matched
// if all selector clauses are contained in the provided podLabels
func podSelectorMatchesPodLabels(podSelector map[string]string, podLabels map[string]string) bool {
	for selectKey, selectValue := range podSelector {
		if labelValue, found := podLabels[selectKey]; found {
			if selectValue != labelValue {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

// CollectorFunction collects all endpoints exposed by pods within host network
// then it collects all unmatched endpoints,
// and then it creates corresponding relations
func (crl *Service2PodCorrelator) CorrelateFunction() error {

	// Services * Pods complexity, per namespace. We can optimize later, but performance should be ok for real world scenarios.
	// Label selectors for services do not cross namespace boundaries
	pods := map[string]([]*PodEndpointCorrelation){}
	for podCorr := range crl.PodCorrChan {
		if podList, found := pods[podCorr.PodNamespace]; found {
			podList = append(podList, podCorr)
			pods[podCorr.PodNamespace] = podList
		} else {
			podList := make([]*PodEndpointCorrelation, 0)
			podList = append(podList, podCorr)
			pods[podCorr.PodNamespace] = podList
		}
	}

	// now for every endpoint we are trying to find a pod serving it
	for svcCorr := range crl.EndpointCorrChannel {
		serviceID := svcCorr.ServiceExternalID

		if namespacePods, found := pods[svcCorr.Namespace]; found {
			for _, pod := range namespacePods {
				if podSelectorMatchesPodLabels(svcCorr.LabelSelector, pod.Labels) {
					podID := crl.buildPodExternalID(pod.PodNamespace, pod.PodName)
					relation := crl.CreateRelation(serviceID, podID, "exposes")
					log.Tracef("Correlated StackState service -> pod relation %s->%s", relation.SourceID, relation.TargetID)
					crl.SubmitRelation(relation)
				}
			}
		}
	}

	return nil
}
