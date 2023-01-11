package features

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/StackVista/stackstate-agent/pkg/httpclient"
	"github.com/stretchr/testify/assert"
)

const validResponse = `{"interpreted-spans":true,"ingest-telemetry-data":true,"max-relation-creates-per-hour-per-agent":5000,"max-component-creates-per-hour-per-agent":2000,"max-connections-per-agent":150000,"upgrade-to-multi-metrics":true,"incremental-topology":true,"expose-kubernetes-status":true,"max-components-per-agent":5000}`

func TestFeaturesAreFetched(t *testing.T) {
	feats := InitTestFeatures(NewRetryableHttpClientStub(validResponse))
	assert.True(t, feats.FeatureEnabled(ExposeKubernetesSatus))
}

type RetryableHTTPClientStub struct {
	getResponse string
}

func NewRetryableHttpClientStub(getResponse string) RetryableHTTPClientStub {
	return RetryableHTTPClientStub{
		getResponse: getResponse,
	}
}

func (h RetryableHTTPClientStub) Get(path string) *httpclient.HTTPResponse {
	return &httpclient.HTTPResponse{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(h.getResponse)),
		},
		Err: nil,
	}
}
func (h RetryableHTTPClientStub) Put(path string, body []byte) *httpclient.HTTPResponse {
	return nil
}
func (h RetryableHTTPClientStub) Post(path string, body []byte) *httpclient.HTTPResponse {
	return nil
}
