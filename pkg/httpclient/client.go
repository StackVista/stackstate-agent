package httpclient

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/trace/info"
	"github.com/StackVista/stackstate-agent/pkg/trace/watchdog"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// GET is used for HTTP GET calls
const GET = "GET"

// POST is used for HTTP POST calls
const POST = "POST"

// PUT is used for HTTP PUT calls
const PUT = "PUT"

// HTTPResponse is used to represent the response from the request
type HTTPResponse struct {
	Response    *http.Response
	Body        []byte
	RetriesLeft int
	Err         error
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
}

// RetryableHTTPClient describes the functionality of a http client with retries and backoff
type RetryableHTTPClient interface {
	Get(path string) *HTTPResponse
	GetWithRetry(path string, retryInterval time.Duration, retryCount int) *HTTPResponse
	Put(path string, body []byte) *HTTPResponse
	PutWithRetry(path string, body []byte, retryInterval time.Duration, retryCount int) *HTTPResponse
	Post(path string, body []byte) *HTTPResponse
	PostWithRetry(path string, body []byte, retryInterval time.Duration, retryCount int) *HTTPResponse
}

// RetryableHTTPClient creates a http client to communicate to StackState
type retryableHTTPClient struct {
	*ClientHost
	*http.Client
	mux sync.Mutex
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
func newClient(host *ClientHost) *http.Client {
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
	return &http.Client{Timeout: 30 * time.Second, Transport: transport}
}

// Get performs a GET request to some path
func (rc *retryableHTTPClient) Get(path string) *HTTPResponse {
	return rc.requestRetryHandler(GET, path, nil, 5*time.Second, 5)
}

// GetWithRetry performs a GET request to some path with a set retry interval and count
func (rc *retryableHTTPClient) GetWithRetry(path string, retryInterval time.Duration, retryCount int) *HTTPResponse {
	return rc.requestRetryHandler(GET, path, nil, retryInterval, retryCount)
}

// Put performs a PUT request to some path
func (rc *retryableHTTPClient) Put(path string, body []byte) *HTTPResponse {
	return rc.requestRetryHandler(PUT, path, body, 5*time.Second, 5)
}

// PutWithRetry performs a PUT request to some path with a set retry interval and count
func (rc *retryableHTTPClient) PutWithRetry(path string, body []byte, retryInterval time.Duration, retryCount int) *HTTPResponse {
	return rc.requestRetryHandler(PUT, path, body, retryInterval, retryCount)
}

// Post performs a POST request to some path
func (rc *retryableHTTPClient) Post(path string, body []byte) *HTTPResponse {
	return rc.requestRetryHandler(POST, path, body, 5*time.Second, 5)
}

// PostWithRetry performs a POST request to some path with a set retry interval and count
func (rc *retryableHTTPClient) PostWithRetry(path string, body []byte, retryInterval time.Duration, retryCount int) *HTTPResponse {
	return rc.requestRetryHandler(POST, path, body, retryInterval, retryCount)
}

func (rc *retryableHTTPClient) requestRetryHandler(method, path string, body []byte, retryInterval time.Duration, retryCount int) *HTTPResponse {
	retryTicker := time.NewTicker(retryInterval)
	retriesLeft := retryCount
	responseChan := make(chan *HTTPResponse, 1)
	waitResponseChan := make(chan *HTTPResponse, 1)

	defer watchdog.LogOnPanic()
	defer close(responseChan)

	go func() {
	retry:
		for {
			select {
			case <-retryTicker.C:
				rc.handleRequest(method, path, body, retriesLeft, responseChan)
				rc.mux.Lock()
				// Lock so we can decrement the retriesLeft
				retriesLeft = retriesLeft - 1
				rc.mux.Unlock()
			case response := <-responseChan:
				// Stop retrying and return the response
				retryTicker.Stop()
				waitResponseChan <- response
				break retry
			}
		}
	}()

	response := <-waitResponseChan
	rc.mux.Lock()
	response.RetriesLeft = retriesLeft
	rc.mux.Unlock()
	return response
}

// getSupportedFeatures returns the features supported by the StackState API
func (rc *retryableHTTPClient) handleRequest(method, path string, body []byte, retriesLeft int, responseChan chan *HTTPResponse) {
	rc.mux.Lock()
	// Lock so only one goroutine at a time can access the map
	if retriesLeft == 0 {
		responseChan <- &HTTPResponse{Err: errors.New("failed after all retries")}
	}
	rc.mux.Unlock()

	response, err := rc.makeRequest(method, path, body)

	// Handle error response
	if err != nil {
		// Soo we got a 404, meaning we were able to contact StackState, but it did not have the requested path. We can publish a result
		if response != nil {
			//responseChan <- &HTTPResponse{
			//	RetriesLeft: retriesLeft,
			//	Err: errors.New("found StackState version which does not have the requested path"),
			//}
			return
		}
		// Log
		_ = log.Error(err)
		return
	}

	defer response.Body.Close()

	// Get byte array
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		_ = log.Errorf("could not decode response body from request: %s", err)
		return
	}

	responseChan <- &HTTPResponse{Response: response, Body: body, Err: nil}
}

// makeRequest
func (rc *retryableHTTPClient) makeRequest(method, path string, body []byte) (*http.Response, error) {
	url := fmt.Sprintf("%s/%s", rc.Host, path)
	var req *http.Request
	var err error
	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("could not create request to %s/%s: %s", url, path, err)
	}

	req.Header.Add("content-encoding", "identity")
	//req.Header.Add("Content-Type", "application/json")
	req.Header.Add("sts-api-key", rc.APIKey)
	req.Header.Add("sts-hostname", rc.Host)
	req.Header.Add("sts-agent-version", info.Version)

	resp, err := rc.Do(req)
	if err != nil {
		if rc.isHTTPTimeout(err) {
			return nil, fmt.Errorf("timeout detected on %s, %s", url, err)
		}
		return nil, fmt.Errorf("error submitting payload to %s: %s", url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		defer resp.Body.Close()
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		return resp, fmt.Errorf("unexpected response from %s. Status: %s", url, resp.Status)
	}

	return resp, nil
}

// IsTimeout returns true if the error is due to reaching the timeout limit on the http.client
func (rc *retryableHTTPClient) isHTTPTimeout(err error) bool {
	if netErr, ok := err.(interface {
		Timeout() bool
	}); ok && netErr.Timeout() {
		return true
	} else if strings.Contains(err.Error(), "use of closed network connection") { //To deprecate when using GO > 1.5
		return true
	}
	return false
}
