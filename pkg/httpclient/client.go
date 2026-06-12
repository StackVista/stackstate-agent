package httpclient

import (
	"compress/gzip"
	pkgconfigsetup "github.com/DataDog/datadog-agent/pkg/config/setup"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/httpclient"
	"net/url"
)

// NewStackStateClient returns a RetryableHTTPClient containing a http.Client configured with the Agent options.
func NewStackStateClient() *httpclient.StackStateClient {
	host := &httpclient.ClientHost{}

	host.HostURL = pkgconfigsetup.Datadog().GetString("sts_url")
	host.APIKey = pkgconfigsetup.Datadog().GetString("api_key")
	host.ContentEncoding = httpclient.NewGzipContentEncoding(gzip.BestCompression)
	host.RetryWaitMin = httpclient.DefaultRetryMin
	host.RetryWaitMax = httpclient.DefaultRetryMax
	host.NoProxy = true

	if addr := pkgconfigsetup.Datadog().GetString("proxy.https"); addr != "" {
		url, err := url.Parse(addr)
		if err == nil {
			host.ProxyURL = url
		} else {
			log.Errorf("Failed to parse proxy URL from proxy.https configuration: %s", err)
		}
	}

	if pkgconfigsetup.Datadog().IsSet("skip_ssl_validation") {
		host.SkipSSLValidation = pkgconfigsetup.Datadog().GetBool("skip_ssl_validation")
	}

	return httpclient.NewStackStateClient(host)
}
