package features

import (
	"encoding/json"
	"github.com/StackVista/stackstate-receiver-go-client/pkg/httpclient"
	"io/ioutil"
	"reflect"

	clientwrapper "github.com/DataDog/datadog-agent/pkg/httpclient"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// FeatureID type ensures well-defined list of features in this file
type FeatureID string

// List of features managed by StackState
const (
	UpgradeToMultiMetrics  FeatureID = "upgrade-to-multi-metrics"
	IncrementalTopology    FeatureID = "incremental-topology"
	HealthStates           FeatureID = "health-states"
	ExposeKubernetesStatus FeatureID = "expose-kubernetes-status"
)

// Features is used to determine whether StackState supports a certain feature or not
type Features interface {
	// FeatureEnabled is used to check whether a feature is enabled
	FeatureEnabled(feature FeatureID) bool
}

type featureSet = map[FeatureID]bool

// FetchFeatures uses the StackState api to retrieve the supported feature set
type FetchFeatures struct {
	features  featureSet
	stsClient httpclient.RetryableHTTPClient
}

// AllFeatures is intended for tests and returns enabled = true for every Feature
type AllFeatures struct{}

// All returns a Features instance with all features enabled. For tests only
func All() *AllFeatures {
	return &AllFeatures{}
}

// FeatureEnabled check
func (f *AllFeatures) FeatureEnabled(_ FeatureID) bool {
	return true
}

/*
InitFeatures will call out to StackState and will intentionally wait for a response.
This ensures that before any checks are run the supported feature set has been determined.
The feature set will not be updated but remains fixed for the runtime of the agent.
This together allows checks to not worry about changes in supported features while running and
avoids potential bugs in stateful checks.

The trade-off is that an upgrade to StackState will only be recognized after an agent restart
*/
func InitFeatures() *FetchFeatures {
	features := &FetchFeatures{
		features:  make(map[FeatureID]bool),
		stsClient: clientwrapper.NewStackStateClient(),
	}

	features.init()

	return features
}

// InitTestFeatures is for tests only and allows injecting a stub for the httpclient
func InitTestFeatures(stsClient httpclient.RetryableHTTPClient) *FetchFeatures {
	features := &FetchFeatures{
		features:  make(map[FeatureID]bool),
		stsClient: stsClient,
	}

	features.init()

	return features
}

func (ff *FetchFeatures) init() {
	features, err := ff.getFeatures()
	if err != nil {
		log.Warnf("Failed to fetch StackState features. Continuing with empty set for StackState feature.")
	}
	ff.features = features
}

// FeatureEnabled check
func (ff *FetchFeatures) FeatureEnabled(feature FeatureID) bool {
	if supported, ok := ff.features[feature]; ok {
		return supported
	}
	return false
}

func (ff *FetchFeatures) getFeatures() (featureSet, error) {
	response := ff.stsClient.Get("features")

	if response.Err != nil {
		return nil, log.Errorf("Failed to fetch StackState features, %s", response.Err)
	}
	if response.Response.StatusCode == 404 {
		log.Info("Found StackState version which does not support feature detection yet")
		return make(map[FeatureID]bool), nil
	}

	if response.Response.StatusCode < 200 || response.Response.StatusCode >= 300 {
		return make(map[FeatureID]bool), log.Errorf("Failed to fetch StackState features, got status code %d", response.Response.StatusCode)
	}

	defer response.Response.Body.Close()

	// Get byte array
	body, err := ioutil.ReadAll(response.Response.Body)
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
