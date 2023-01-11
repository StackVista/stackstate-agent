package features

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"

	coreconfig "github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/StackVista/stackstate-agent/pkg/util/retry"
)

// FeatureID type ensures well-defined list of features in this file
type FeatureID string

// List of features managed by StackState receiver
const (
	UpgradeToMultiMetrics FeatureID = "upgrade-to-multi-metrics"
	IncrementalTopology   FeatureID = "incremental-topology"
	HealthStates          FeatureID = "health-states"
	ExposeKubernetesSatus FeatureID = "expose-kubernetes-status"
)

type Features interface {
	FeatureEnabled(feature FeatureID) bool
}

type featureSet = map[FeatureID]bool

type FetchFeatures struct {
	features   featureSet
	httpClient http.Client
	endpoint   string
	apiKey     string
}

type AllFeatures struct{}

func All() *AllFeatures {
	return &AllFeatures{}
}

// FeatureEnabled check
func (f *AllFeatures) FeatureEnabled(_ FeatureID) bool {
	return true
}

// TODO use some existing HTTP client or generic client to interact with stackstate api
func InitFeatures() *FetchFeatures {
	mainEndpoint := coreconfig.GetMainInfraEndpoint()
	apiKeyPerEndpoint, err := coreconfig.GetMultipleEndpoints()

	mainAPIKey := ""
	if err != nil {
		log.Warnf("Failed to fetch StackState features, no API key configured. Continuing with empty set for StackState.")
	} else {
		mainAPIKey = apiKeyPerEndpoint[mainEndpoint][0]
	}

	features := &FetchFeatures{
		features:   make(map[FeatureID]bool),
		endpoint:   mainEndpoint,
		apiKey:     mainAPIKey,
		httpClient: http.Client{Timeout: 30 * time.Second}, //, Transport: cfg.Transport},
	}

	if err == nil {
		if features.init() != nil {
			log.Warnf("Failed to fetch StackState features. Continuing with empty set for StackState feature.")
		}
	}

	return features
}

func (af *FetchFeatures) init() error {
	initRetry := retry.Retrier{}
	initRetry.SetupRetrier(&retry.Config{ //nolint:errcheck
		Name: "FechFeaturesFromStackState",
		AttemptMethod: func() error {
			features, err := af.getFeatures()
			af.features = features
			return err
		},
		Strategy:   retry.RetryCount,
		RetryDelay: 1 * time.Second,
		RetryCount: 5,
	})

	var err error
	for {
		err = initRetry.TriggerRetry()
		isRetryError, retryError := retry.IsRetryError(err)
		if !isRetryError {
			return err
		}
		if retryError.RetryStatus == retry.PermaFail {
			return retryError.Unwrap()
		}
		time.Sleep(time.Nanosecond)
	}
}

func (f *FetchFeatures) FeatureEnabled(feature FeatureID) bool {
	if supported, ok := f.features[feature]; ok {
		return supported
	}
	return false
}

// func (af *FetchFeatures) Stop() {
// 	af.stop <- true
// }

// Will try to fetch features from endpoint until it succeeds once
// TODO: Shouldn't it keep fetching features to detect changes (in case StackState is upgraded for example)?
// func (af *FetchFeatures) run() {
// 	featuresTicker := time.NewTicker(5 * time.Second)
// 	// Channel to announce new features detected
// 	featuresCh := make(chan featureSet, 1)

// 	af.getFeaturesAsync(featuresCh)

// 	go func() {
// 		for {
// 			select {
// 			case <-featuresTicker.C:
// 				go af.getFeaturesAsync(featuresCh)
// 			case featuresValue := <-featuresCh:
// 				af.features = featuresValue
// 				// Stop polling
// 				featuresTicker.Stop()
// 				// case <-af.stop:
// 				// 	return
// 			}
// 		}
// 	}()
// }

func (af *FetchFeatures) getFeaturesAsync(featuresCh chan featureSet) {
	features, err := af.getFeatures()

	if err != nil {
		// Ignoring errors, they are already logged
		return
	}
	featuresCh <- features
}

func (af *FetchFeatures) getFeatures() (featureSet, error) {
	resp, accessErr := af.accessAPIwithEncoding("GET", "/features", make([]byte, 0), "identity")

	// Handle error response
	if accessErr != nil {
		// So we got a 404, meaning we were able to contact stackstate, but it had no features path. We can publish a result
		if resp != nil {
			log.Info("Found StackState version which does not support feature detection yet")
			return make(map[FeatureID]bool), nil
		}
		// Log
		return nil, log.Errorf("Failed to fetch StackState features, %s", accessErr)
	}

	defer resp.Body.Close()

	// Get byte array
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, log.Errorf("Failed to fetch features, error while decoding response: %s", err)
	}
	var data interface{}
	// Parse json
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, log.Errorf("Failed to fetch features, error while unmarshalling response: %s of body %s", err, body)
	}

	// Validate structure
	featureMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, log.Errorf("Failed to fetch features, json was wrongly formatted, expected map type, got: %s", reflect.TypeOf(data))
	}

	featuresParsed := make(map[FeatureID]bool)

	for k, v := range featureMap {
		featureValue, okV := v.(bool)
		if !okV {
			_ = log.Warnf("Fetching features found wrong type in json response, expected boolean type, got: %s, skipping feature %s", reflect.TypeOf(v), k)
		}
		featuresParsed[FeatureID(k)] = featureValue
	}

	log.Infof("Fetching features, server supports features: %s", featuresParsed)
	return featuresParsed, nil
}

func (af *FetchFeatures) accessAPIwithEncoding(method string, checkPath string, body []byte, contentEncoding string) (*http.Response, error) {
	url := fmt.Sprintf("%s?APIKey=%s", af.endpoint, af.apiKey) + checkPath // Add the checkPath in full Process Agent URL
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("could not create %s request to %s: %s", method, url, err)
	}

	req.Header.Add("content-encoding", contentEncoding)
	req.Header.Add("sts-api-key", af.apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), ReqCtxTimeout)
	defer cancel()
	req.WithContext(ctx)

	resp, err := af.httpClient.Do(req)
	if err != nil {
		if isHTTPTimeout(err) {
			return nil, fmt.Errorf("Timeout detected on %s, %s", url, err)
		}
		return nil, fmt.Errorf("Error submitting payload to %s: %s", url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		defer resp.Body.Close()
		io.Copy(ioutil.Discard, resp.Body)
		return resp, fmt.Errorf("unexpected response from %s. Status: %s, Body: %v", url, resp.Status, resp.Body)
	}

	return resp, nil

}

const (
	// HTTPTimeout is the timeout in seconds for process-agent to send process payloads to DataDog
	HTTPTimeout = 20 * time.Second
	// ReqCtxTimeout is the timeout in seconds for process-agent to cancel POST request using context timeout
	ReqCtxTimeout = 30 * time.Second
)

// IsTimeout returns true if the error is due to reaching the timeout limit on the http.client
func isHTTPTimeout(err error) bool {
	if netErr, ok := err.(interface {
		Timeout() bool
	}); ok && netErr.Timeout() {
		return true
	} else if strings.Contains(err.Error(), "use of closed network connection") { //To deprecate when using GO > 1.5
		return true
	}
	return false
}
