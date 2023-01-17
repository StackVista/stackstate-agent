package features

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/StackVista/stackstate-agent/pkg/httpclient"
	"github.com/stretchr/testify/assert"
)

const validResponse = `{"interpreted-spans":true,"ingest-telemetry-data":true,"max-relation-creates-per-hour-per-agent":5000,"max-component-creates-per-hour-per-agent":2000,"max-connections-per-agent":150000,"upgrade-to-multi-metrics":true,"incremental-topology":true,"expose-kubernetes-status":true,"max-components-per-agent":5000}`
const invalidResponse = `"interpreted-spans":true,"ingest-telemetry-data":true,"max-relation-creates-per-hour-per-agent":5000,"max-component-creates-per-hour-per-agent":2000,"max-connections-per-agent":150000,"upgrade-to-multi-metrics":true,"incremental-topology":true,"expose-kubernetes-status":true,"max-components-per-agent":5000}`

func TestFeaturesAreFetched(t *testing.T) {
	feats := InitTestFeatures(newRetryableHTTPClientStub(createHTTPResponse(validResponse, 200, nil)))
	assert.True(t, feats.FeatureEnabled(ExposeKubernetesStatus))
	assert.True(t, feats.FeatureEnabled(UpgradeToMultiMetrics))
	assert.True(t, feats.FeatureEnabled(IncrementalTopology))
	assert.False(t, feats.FeatureEnabled(HealthStates))
}

func TestInvalidResponseContinuousWithNoFeaturesEnabled(t *testing.T) {
	feats := InitTestFeatures(newRetryableHTTPClientStub(createHTTPResponse(invalidResponse, 200, nil)))
	assert.False(t, feats.FeatureEnabled(ExposeKubernetesStatus))
	assert.False(t, feats.FeatureEnabled(UpgradeToMultiMetrics))
	assert.False(t, feats.FeatureEnabled(IncrementalTopology))
	assert.False(t, feats.FeatureEnabled(HealthStates))
}

func Test404ResponseContinousWithNoFeaturesEnabled(t *testing.T) {
	feats := InitTestFeatures(newRetryableHTTPClientStub(createHTTPResponse("", 404, nil)))
	assert.False(t, feats.FeatureEnabled(ExposeKubernetesStatus))
	assert.False(t, feats.FeatureEnabled(UpgradeToMultiMetrics))
	assert.False(t, feats.FeatureEnabled(IncrementalTopology))
	assert.False(t, feats.FeatureEnabled(HealthStates))
}

func TestErrorContinousWithNoFeaturesEnabled(t *testing.T) {
	feats := InitTestFeatures(newRetryableHTTPClientStub(createHTTPResponse("", 0, errors.New("Failed http call"))))
	assert.False(t, feats.FeatureEnabled(ExposeKubernetesStatus))
	assert.False(t, feats.FeatureEnabled(UpgradeToMultiMetrics))
	assert.False(t, feats.FeatureEnabled(IncrementalTopology))
	assert.False(t, feats.FeatureEnabled(HealthStates))
}

func TestHttpErrorContinousWithNoFeaturesEnabled(t *testing.T) {
	feats := InitTestFeatures(newRetryableHTTPClientStub(createHTTPResponse("", 400, nil)))
	assert.False(t, feats.FeatureEnabled(ExposeKubernetesStatus))
	assert.False(t, feats.FeatureEnabled(UpgradeToMultiMetrics))
	assert.False(t, feats.FeatureEnabled(IncrementalTopology))
	assert.False(t, feats.FeatureEnabled(HealthStates))
}

type RetryableHTTPClientStub struct {
	response *httpclient.HTTPResponse
}

func newRetryableHTTPClientStub(response *httpclient.HTTPResponse) RetryableHTTPClientStub {
	return RetryableHTTPClientStub{
		response: response,
	}
}

func (h RetryableHTTPClientStub) Get(path string) *httpclient.HTTPResponse {
	return h.response
}
func (h RetryableHTTPClientStub) Put(path string, body []byte) *httpclient.HTTPResponse {
	return h.response
}
func (h RetryableHTTPClientStub) Post(path string, body []byte) *httpclient.HTTPResponse {
	return h.response
}

func createHTTPResponse(body string, statusCode int, err error) *httpclient.HTTPResponse {
	return &httpclient.HTTPResponse{
		Response: &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(strings.NewReader(body)),
		},
		Err: err,
	}
}
