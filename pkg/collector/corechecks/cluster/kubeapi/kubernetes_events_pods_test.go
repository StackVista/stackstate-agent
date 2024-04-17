//go:build kubeapiserver

package kubeapi

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/DataDog/datadog-agent/pkg/aggregator/mocksender"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

func TestProcessPodEvents(t *testing.T) {
	rawPodEvents, _ := ioutil.ReadFile("./testdata/process_pod_events_data.json")

	var podEvents []*v1.Pod
	err := json.Unmarshal(rawPodEvents, &podEvents)
	if err != nil {
		t.Fatal("Unable to case process_pod_events_data.json into []*v1.Event")
	}

	evCheck := KubernetesAPIEventsFactory().(*EventsCheck)
	evCheck.ac = MockAPIClient(nil)

	mockSender := mocksender.NewMockSender(evCheck.ID())
	mockSender.On("Event", mock.AnythingOfType("metrics.Event"))

	evCheck.processPods(mockSender, podEvents)
	mockSender.AssertNumberOfCalls(t, "Event", 1)
}
