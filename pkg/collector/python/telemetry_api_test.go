package python

import "testing"

func TestTopologyEvent(t *testing.T) {
	testTopologyEvent(t)
}

func TestTopologyEventMissingFields(t *testing.T) {
	testTopologyEventMissingFields(t)
}

func TestTopologyEventWrongFieldType(t *testing.T) {
	testTopologyEventWrongFieldType(t)
}

func TestSubmitRawMetricsData(t *testing.T) {
	testRawMetricsData(t)
}
