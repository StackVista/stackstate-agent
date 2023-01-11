package features

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"time"

	"github.com/StackVista/stackstate-agent/pkg/httpclient"
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
	features  featureSet
	stsClient httpclient.RetryableHTTPClient
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
	features := &FetchFeatures{
		features:  make(map[FeatureID]bool),
		stsClient: httpclient.NewStackStateClient(),
	}

	if features.init() != nil {
		log.Warnf("Failed to fetch StackState features. Continuing with empty set for StackState feature.")
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
		if err == nil {
			return nil
		}
		if isRetryError, retryError := retry.IsRetryError(err); isRetryError {
			if retryError.RetryStatus == retry.PermaFail {
				return retryError.Unwrap()
			}
		} else {
			return err
		}
		time.Sleep(500 * time.Microsecond)
	}
}

func (f *FetchFeatures) FeatureEnabled(feature FeatureID) bool {
	if supported, ok := f.features[feature]; ok {
		return supported
	}
	return false
}

func (af *FetchFeatures) getFeaturesAsync(featuresCh chan featureSet) {
	features, err := af.getFeatures()

	if err != nil {
		// Ignoring errors, they are already logged
		return
	}
	featuresCh <- features
}

func (af *FetchFeatures) getFeatures() (featureSet, error) {
	response := af.stsClient.Get("/features")

	if response.Response.StatusCode == 404 {
		log.Info("Found StackState version which does not support feature detection yet")
		return make(map[FeatureID]bool), nil
	}
	if response.Err != nil {
		return nil, log.Errorf("Failed to fetch StackState features, %s", response.Err)
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
