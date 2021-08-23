// +build python,test

// TODO: Raw Metrics

package python

import "testing"

func TestSubmitRawMetricsData(t *testing.T) {
	testHealthCheckData(t)
}

func TestSubmitRawMetricsStartSnapshot(t *testing.T) {
	testHealthStartSnapshot(t)
}

func TestSubmitRawMetricsStopSnapshot(t *testing.T) {
	testHealthStopSnapshot(t)
}

func TestNoRawMetricsSubStream(t *testing.T) {
	testNoSubStream(t)
}
