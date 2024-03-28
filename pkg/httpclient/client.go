package httpclient

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/StackVista/stackstate-agent/pkg/version"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

// GET is used for HTTP GET calls
const GET = "GET"

// POST is used for HTTP POST calls
const POST = "POST"

// PUT is used for HTTP PUT calls
const PUT = "PUT"

// HTTPResponse is used to represent the response from the request
type HTTPResponse struct {
	Response       *http.Response
	RequestPayload []byte
	Err            error
}

// ClientHost specifies an host that the client communicates with.
type ClientHost struct {
	APIKey string `json:"-"` // never marshal this
	Host   string

	// NoProxy will be set to true when the proxy setting for the trace API endpoint
	// needs to be ignored (e.g. it is part of the "no_proxy" list in the yaml settings).
	NoProxy           bool
	ProxyURL          *url.URL
	SkipSSLValidation bool

	ContentEncoding ContentEncoding // TODO: make this per request
}

// RetryableHTTPClient describes the functionality of a http client with retries and backoff
type RetryableHTTPClient interface {
	Get(path string) *HTTPResponse
	Put(path string, body []byte) *HTTPResponse
	Post(path string, body []byte) *HTTPResponse
}

// RetryableHTTPClient creates a http client to communicate to StackState
type retryableHTTPClient struct {
	*ClientHost
	*retryablehttp.Client
}

// StackStateClient creates a wrapper around the RetryableHTTPClient that is used for communication with StackState over http(s)
type StackStateClient struct {
	RetryableHTTPClient
}

// NewStackStateClient returns a RetryableHTTPClient containing a http.Client configured with the Agent options.
func NewStackStateClient() RetryableHTTPClient {
	return &StackStateClient{NewHTTPClient("sts_url")}
}

// NewHTTPClient returns a RetryableHTTPClient containing a http.Client configured with the Agent options.
func NewHTTPClient(baseURLConfigKey string) RetryableHTTPClient {
	return makeRetryableHTTPClient(baseURLConfigKey)
}

func makeRetryableHTTPClient(baseURLConfigKey string) RetryableHTTPClient {
	host := &ClientHost{}
	if hostURL := config.Datadog.GetString(baseURLConfigKey); hostURL != "" {
		host.Host = hostURL
	}

	host.APIKey = config.Datadog.GetString("api_key")

	host.ContentEncoding = NewGzipContentEncoding(gzip.BestCompression)

	proxyList := config.Datadog.GetStringSlice("proxy.no_proxy")
	noProxy := make(map[string]bool, len(proxyList))
	for _, host := range proxyList {
		// map of hosts that need to be skipped by proxy
		noProxy[host] = true
	}
	host.NoProxy = noProxy[host.Host]

	if addr := config.Datadog.GetString("proxy.https"); addr != "" {
		url, err := url.Parse(addr)
		if err == nil {
			host.ProxyURL = url
		} else {
			log.Errorf("Failed to parse proxy URL from proxy.https configuration: %s", err)
		}
	}

	if config.Datadog.IsSet("skip_ssl_validation") {
		host.SkipSSLValidation = config.Datadog.GetBool("skip_ssl_validation")
	}

	return &retryableHTTPClient{
		ClientHost: host,
		Client:     newClient(host),
	}
}

// newClient returns a http.Client configured with the Agent options.
func newClient(host *ClientHost) *retryablehttp.Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: host.SkipSSLValidation},
	}
	if host.ProxyURL != nil && !host.NoProxy {
		log.Infof("configuring proxy through: %s", host.ProxyURL.String())
		transport.Proxy = http.ProxyURL(host.ProxyURL)
	}

	retryableClient := retryablehttp.NewClient()
	// Make retryableClient logging use log level settings by replace the default logger with wrapped leveledLogger
	retryableClient.Logger = &leveledLogger{}
	retryableClient.HTTPClient = &http.Client{Timeout: 30 * time.Second, Transport: transport}
	if config.Datadog.IsSet("transactional_forwarder_retry_min") {
		retryableClient.RetryWaitMin = config.Datadog.GetDuration("transactional_forwarder_retry_min")
	}

	if config.Datadog.IsSet("transactional_forwarder_retry_max") {
		retryableClient.RetryWaitMax = config.Datadog.GetDuration("transactional_forwarder_retry_max")
	}

	return retryableClient
}

// Get performs a GET request to some path
func (rc *retryableHTTPClient) Get(path string) *HTTPResponse {
	return rc.handleRequest(GET, path, nil)
}

// Put performs a PUT request to some path
func (rc *retryableHTTPClient) Put(path string, body []byte) *HTTPResponse {
	return rc.handleRequest(PUT, path, body)
}

// Post performs a POST request to some path
func (rc *retryableHTTPClient) Post(path string, body []byte) *HTTPResponse {
	return rc.handleRequest(POST, path, body)
}

// getSupportedFeatures returns the features supported by the StackState API
func (rc *retryableHTTPClient) handleRequest(method, path string, body []byte) *HTTPResponse {

	req, err := rc.makeRequest(method, path, body)
	if err != nil {
		return &HTTPResponse{
			Err: err,
		}
	}
	response, err := rc.Do(req)

	return &HTTPResponse{Response: response, RequestPayload: body, Err: err}
}

// makeRequest
func (rc *retryableHTTPClient) makeRequest(method, path string, body []byte) (*retryablehttp.Request, error) {
	url := fmt.Sprintf("%s/%s%s", rc.Host, path, fmt.Sprintf("?api_key=%s", rc.APIKey))
	var req *retryablehttp.Request
	var err error
	if body != nil {
		if gzipped, encodingError := rc.ContentEncoding.encode(body); encodingError != nil {
			_ = log.Warnf("http client was not able to send payload as %s, reverting to uncompressed payload: %s",
				rc.ContentEncoding.name(), err)
			req, err = retryablehttp.NewRequest(method, url, bytes.NewBuffer(body))
		} else {
			log.Debugf("Using %s compression for payload", rc.ContentEncoding.name())
			req, err = retryablehttp.NewRequest(method, url, gzipped)
			req.Header.Add("content-encoding", rc.ContentEncoding.name())
		}
	} else {
		req, err = retryablehttp.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("could not create request to %s/%s: %s", url, path, err)
	}

	req.Header.Add("sts-api-key", rc.APIKey)
	req.Header.Add("sts-hostname", rc.Host)
	req.Header.Add("sts-agent-version", version.AgentVersion)
	req.Header.Add("content-type", "application/json")

	return req, nil
}

// Wrapper for the agent logger to be recognized as a retryablehttp.LeveledLogger
type leveledLogger struct {
}

func (l *leveledLogger) Error(msg string, keysAndValues ...interface{}) {
	log.Errorf(msg, keysAndValues...)
}
func (l *leveledLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Infof(msg, keysAndValues...)
}
func (l *leveledLogger) Debug(msg string, keysAndValues ...interface{}) {
	log.Debugf(msg, keysAndValues)
}
func (l *leveledLogger) Warn(msg string, keysAndValues ...interface{}) {
	log.Warn(msg, keysAndValues)
}
