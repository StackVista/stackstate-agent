package httpclient

import (
	"compress/gzip"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/httpclient"
	"net/url"
)

// NewStackStateClient returns a RetryableHTTPClient containing a http.Client configured with the Agent options.
func NewStackStateClient() *httpclient.StackStateClient {
	host := &httpclient.ClientHost{}

	host.HostURL = config.Datadog.GetString("sts_url")
	host.APIKey = config.Datadog.GetString("api_key")
	host.ContentEncoding = httpclient.NewGzipContentEncoding(gzip.BestCompression)
	host.RetryWaitMin = httpclient.DefaultRetryMin
	host.RetryWaitMax = httpclient.DefaultRetryMax
	host.NoProxy = true

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

	return httpclient.NewStackStateClient(host)
}
