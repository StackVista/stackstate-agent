//go:build kubeapiserver

package topologycollectors

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/DataDog/datadog-agent/pkg/topology"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	v1 "k8s.io/api/core/v1"
)

const redactedMessage string = "<redacted>"

var annotationsToOfuscate = [...]string{"kubectl.kubernetes.io/last-applied-configuration", "openshift.io/token-secret.value"}

// SecretCollector implements the ClusterTopologyCollector interface.
type SecretCollector struct {
	ClusterTopologyCollector
}

// NewSecretCollector creates a new instance of the secret collector
func NewSecretCollector(clusterTopologyCollector ClusterTopologyCollector) ClusterTopologyCollector {
	return &SecretCollector{
		ClusterTopologyCollector: clusterTopologyCollector,
	}
}

// GetName returns the name of the Collector
func (*SecretCollector) GetName() string {
	return "Secret Collector"
}

// CollectorFunction Collects and Published the Secret Components
func (cmc *SecretCollector) CollectorFunction() error {
	secrets, err := cmc.GetAPIClient().GetSecrets()
	if err != nil {
		return err
	}

	for _, cm := range secrets {
		comp, err := cmc.secretToStackStateComponent(cm)
		if err != nil {
			return err
		}

		cmc.SubmitComponent(comp)
	}

	return nil
}

// Creates a StackState Secret component from a Kubernetes / OpenShift Cluster
func (cmc *SecretCollector) secretToStackStateComponent(secret v1.Secret) (*topology.Component, error) {
	log.Tracef("Mapping Secret to StackState component: %s", secret.String())

	// k8s object TypeMeta seem to be archived, it's always empty.
	tags := cmc.initTags(secret.ObjectMeta, metav1.TypeMeta{Kind: "Secret"})
	secretExternalID := cmc.buildSecretExternalID(secret.Namespace, secret.Name)

	// update all annotations that could lead to secrets leak
	for _, annotationName := range annotationsToOfuscate {
		if _, ok := secret.Annotations[annotationName]; ok {
			secret.Annotations[annotationName] = redactedMessage
		}
	}
	secretDataHash, err := secure(secret.Data)
	if err != nil {
		return nil, err
	}

	certExpiration := time.Time{}
	if secret.Type == corev1.SecretTypeTLS {
		if v, ok := secret.Data[corev1.TLSCertKey]; ok {
			certExpiration = certificateExpiration(v)
		} else {
			log.Debugf("TLS Secret %s does not contain a TLS certificate", secretExternalID)
		}
	}

	prunedSecret := secret
	prunedSecret.Data = map[string][]byte{
		"<data hash>": []byte(secretDataHash),
	}

	component := &topology.Component{
		ExternalID: secretExternalID,
		Type:       topology.Type{Name: "secret"},
		Data: map[string]interface{}{
			"name":        secret.Name,
			"tags":        tags,
			"identifiers": []string{secretExternalID},
		},
	}

	if !certExpiration.IsZero() {
		component.Data.PutNonEmpty("certificateExpiration", toUnixMilli(certExpiration))
	}

	if cmc.IsSourcePropertiesFeatureEnabled() {
		var sourceProperties map[string]interface{}
		if cmc.IsExposeKubernetesStatusEnabled() {
			sourceProperties = makeSourcePropertiesFullDetails(&prunedSecret)
		} else {
			sourceProperties = makeSourceProperties(&prunedSecret)
		}
		component.SourceProperties = sourceProperties
	} else {
		component.Data.PutNonEmpty("creationTimestamp", secret.CreationTimestamp)
		component.Data.PutNonEmpty("uid", secret.UID)
		component.Data.PutNonEmpty("generateName", secret.GenerateName)
		component.Data.PutNonEmpty("kind", secret.Kind)
		component.Data.PutNonEmpty("data", secretDataHash)
	}

	log.Tracef("Created StackState Secret component %s: %v", secretExternalID, component.JSONString())

	return component, nil
}

func certificateExpiration(certData []byte) time.Time {
	certString := string(certData)

	// It seems that the certificate is not (always) base64 encoded, so we need to check if it is
	if !strings.HasPrefix(certString, "-----BEGIN CERTIFICATE-----") {
		cd, err := base64.StdEncoding.DecodeString(certString)
		if err != nil {
			log.Errorf("Failed to decode TLS certificate data: %s", err)
			return time.Time{}
		}
		certData = cd
	}

	certs, err := DecodeX509CertificateChainBytes(certData)
	if err != nil {
		log.Errorf("Failed to parse TLS certificate: %s", err)
		return time.Time{}
	}

	cert := certs[0]

	return cert.NotAfter
}

// DecodeX509CertificateChainBytes will decode a PEM encoded x509 Certificate chain.
func DecodeX509CertificateChainBytes(certBytes []byte) ([]*x509.Certificate, error) {
	certs := []*x509.Certificate{}

	var block *pem.Block

	for {
		// decode the tls certificate pem
		block, certBytes = pem.Decode(certBytes)
		if block == nil {
			break
		}

		// parse the tls certificate
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing certificate: %s", err)
		}
		certs = append(certs, cert)
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found in the chain")
	}

	return certs, nil
}

func secure(data map[string][]byte) (string, error) {
	hash := sha256.New()
	if len(data) == 0 {
		return hex.EncodeToString(hash.Sum(nil)), nil
	}

	k := keys(data)
	sort.Strings(k) // Sort so that we have a stable hash

	for _, key := range k {
		if _, err := hash.Write([]byte(key)); err != nil {
			return "", err
		}

		val := data[key]
		if _, err := hash.Write(val); err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func keys(data map[string][]byte) []string {
	keys := make([]string, len(data))
	i := 0

	for k := range data {
		keys[i] = k
		i++
	}

	return keys
}

// toUnixMilli converts a time.Time to milliseconds since epoch, as time.UnixMilli() is not available in go 1.16
func toUnixMilli(t time.Time) int64 {
	return t.Unix() * 1000
}
