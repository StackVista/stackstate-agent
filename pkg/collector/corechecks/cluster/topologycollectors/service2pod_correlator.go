//go:build kubeapiserver
// +build kubeapiserver

package topologycollectors

import (
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// Service2PodCorrelator
type Service2PodCorrelator struct {
	PodCorrChan         <-chan *PodLabelCorrelation
	SelectorCorrChannel <-chan *ServiceSelectorCorrelation
	ClusterTopologyCorrelator
}

// PodLabelCorrelation labels for a pod
type PodLabelCorrelation struct {
	Labels       map[string]string
	PodNamespace string
	PodName      string
}

// ServiceSelectorCorrelation the selector for a service
type ServiceSelectorCorrelation struct {
	ServiceExternalID string
	Namespace         string
	// Only labelMatchers, no labelExpressions for service selection
	LabelSelector map[string]string
}

// NewService2PodCorrelator creates correlator creates relation from service to pod
// in case where service points to kubernetes host and pod is exposed within the host's network
func NewService2PodCorrelator(
	podCorrChannel chan *PodLabelCorrelation,
	serviceCorrChannel chan *ServiceSelectorCorrelation,
	clusterTopologyCorrelator ClusterTopologyCorrelator,
) ClusterTopologyCorrelator {
	return &Service2PodCorrelator{
		PodCorrChan:               podCorrChannel,
		SelectorCorrChannel:       serviceCorrChannel,
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
	// For services without podSelector, we for now do no matches (see https://stackoverflow.com/a/61866213)
	// We could model this based on endpoints, but that will breaks metric aggregation over services (which
	// is done based on selectors for time travelling reasons). For now this is just something we do not support.
	if len(podSelector) == 0 {
		return false
	}

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

// CollectorFunction collects all pods with namespace and labels,
// then it goes through all services, selecting which service connects to which pod.
func (crl *Service2PodCorrelator) CorrelateFunction() error {

	// Services * Pods complexity, per namespace. We can optimize later, but performance should be ok for real world scenarios.
	// Label selectors for services do not cross namespace boundaries
	pods := map[string]([]*PodLabelCorrelation){}
	for podCorr := range crl.PodCorrChan {
		if podList, found := pods[podCorr.PodNamespace]; found {
			podList = append(podList, podCorr)
			pods[podCorr.PodNamespace] = podList
		} else {
			podList := make([]*PodLabelCorrelation, 0)
			podList = append(podList, podCorr)
			pods[podCorr.PodNamespace] = podList
		}
	}

	// now for every service we are trying to find a pods serving it
	for svcCorr := range crl.SelectorCorrChannel {
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
