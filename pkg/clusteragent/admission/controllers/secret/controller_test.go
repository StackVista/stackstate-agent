// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver
// +build kubeapiserver

package secret

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/StackVista/stackstate-agent/pkg/util/kubernetes/certificate"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

var (
	cfg = Config{
		ns:   "foo",
		name: "bar",
		svc:  "baz",
		cert: CertConfig{
			expirationThreshold: 30 * 24 * time.Hour,
			validityBound:       365 * 24 * time.Hour,
		},
	}
)

func TestCreateSecret(t *testing.T) {
	f := newFixture(t)
	c := f.run(t)

	// Validate that a fresh Secret has been created
	secret, err := c.secretsLister.Secrets(cfg.GetNs()).Get(cfg.GetName())
	if err != nil {
		t.Fatalf("Failed to get the Secret: %v", err)
	}

	if err := validate(secret); err != nil {
		t.Fatalf("Invalid Secret: %v", err)
	}

	if c.queue.Len() != 0 {
		t.Fatal("Work queue isn't empty")
	}
}

func TestRefreshNotRequired(t *testing.T) {
	f := newFixture(t)

	// Create a Secret with a valid certificate
	data, err := certificate.GenerateSecretData(time.Now(), time.Now().Add(365*24*time.Hour), generateDNSNames(cfg.GetNs(), cfg.GetSvc()))
	if err != nil {
		t.Fatalf("Failed to create the Secret: %v", err)
	}

	oldSecret := buildSecret(data)
	f.populateCache(oldSecret)

	c := f.run(t)

	// Validate that the Secret hasn't changed
	newSecret, err := c.secretsLister.Secrets(cfg.GetNs()).Get(cfg.GetName())
	if err != nil {
		t.Fatalf("Failed to get the Secret: %v", err)
	}

	if !reflect.DeepEqual(oldSecret, newSecret) {
		t.Fatal("The Secret has been modified")
	}

	if err := validate(newSecret); err != nil {
		t.Fatalf("Invalid Secret: %v", err)
	}

	if c.queue.Len() != 0 {
		t.Fatal("Work queue isn't empty")
	}
}

func TestRefreshExpiration(t *testing.T) {
	f := newFixture(t)

	// Create a Secret with a certificate expiring soon
	data, err := certificate.GenerateSecretData(time.Now(), time.Now().Add(5*time.Minute), generateDNSNames(cfg.GetNs(), cfg.GetSvc()))
	if err != nil {
		t.Fatalf("Failed to create the Secret: %v", err)
	}

	oldSecret := buildSecret(data)
	f.populateCache(oldSecret)

	c := f.run(t)

	// Validate that the Secret has been refreshed
	newSecret, err := c.secretsLister.Secrets(cfg.GetNs()).Get(cfg.GetName())
	if err != nil {
		t.Fatalf("Failed to get the Secret: %v", err)
	}

	if reflect.DeepEqual(oldSecret, newSecret) {
		t.Fatalf("The Secret hasn't been modified")
	}

	if err := validate(newSecret); err != nil {
		t.Fatalf("Invalid Secret: %v", err)
	}

	if c.queue.Len() != 0 {
		t.Fatal("Work queue isn't empty")
	}
}

func TestRefreshDNSNames(t *testing.T) {
	f := newFixture(t)

	// Create a Secret with a dns name that doesn't match the config
	data, err := certificate.GenerateSecretData(time.Now(), time.Now().Add(365*24*time.Hour), []string{"outdated"})
	if err != nil {
		t.Fatalf("Failed to create the Secret: %v", err)
	}

	oldSecret := buildSecret(data)
	f.populateCache(oldSecret)

	c := f.run(t)

	// Validate that the Secret has been refreshed
	newSecret, err := c.secretsLister.Secrets(cfg.GetNs()).Get(cfg.GetName())
	if err != nil {
		t.Fatalf("Failed to get the Secret: %v", err)
	}

	if reflect.DeepEqual(oldSecret, newSecret) {
		t.Fatalf("The Secret hasn't been modified")
	}

	if err := validate(newSecret); err != nil {
		t.Fatalf("Invalid Secret: %v", err)
	}

	if c.queue.Len() != 0 {
		t.Fatal("Work queue isn't empty")
	}
}

func validate(s *corev1.Secret) error {
	cert, err := certificate.GetCertFromSecret(s.Data)
	if err != nil {
		return err
	}

	expiration := certificate.GetDurationBeforeExpiration(cert)
	if expiration < 364*24*time.Hour {
		return fmt.Errorf("The certificate expires too soon: %v", expiration)
	}
	return nil
}

type fixture struct {
	t      *testing.T
	client *fake.Clientset
}

func newFixture(t *testing.T) *fixture {
	return &fixture{
		t:      t,
		client: fake.NewSimpleClientset(),
	}
}

func (f *fixture) run(t *testing.T) *Controller {
	stopCh := make(chan struct{})
	defer close(stopCh)

	factory := informers.NewSharedInformerFactory(f.client, time.Duration(0))
	c := NewController(
		f.client,
		factory.Core().V1().Secrets(),
		func() bool { return true },
		make(chan struct{}),
		cfg,
	)

	factory.Start(stopCh)
	go c.Run(stopCh)

	// Wait for controller to start watching resources effectively and handling objects
	// before returning it.
	// Otherwise tests will start making assertions before the reconciliation is done.
	lastChange := time.Now()
	lastCount := 0
	for {
		time.Sleep(1 * time.Second)
		count := len(f.client.Actions())
		if count > lastCount {
			lastChange = time.Now()
			lastCount = count
		} else if time.Since(lastChange) > 2*time.Second {
			break
		}
	}

	return c
}

func (f *fixture) populateCache(secrets ...*corev1.Secret) {
	for _, s := range secrets {
		_, _ = f.client.CoreV1().Secrets(s.Namespace).Create(context.TODO(), s, metav1.CreateOptions{})
	}
}

func buildSecret(data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cfg.GetNs(),
			Name:      cfg.GetName(),
		},
		Data: data,
	}
}

func TestDigestDNSNames(t *testing.T) {
	tests := []struct {
		name     string
		dnsNames []string
		want     uint64
	}{
		{
			name:     "nominal case",
			dnsNames: []string{"foo", "bar"},
			want:     12531106902390217800,
		},
		{
			name:     "different order, same digest",
			dnsNames: []string{"bar", "foo"},
			want:     12531106902390217800,
		},
		{
			name:     "empty dnsNames",
			dnsNames: []string{},
			want:     14695981039346656037,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := make([]string, len(tt.dnsNames))
			copy(tmp, tt.dnsNames)

			got := digestDNSNames(tt.dnsNames)
			assert.Equal(t, tt.want, got)

			// Assert we didn't mutate the input
			assert.Equal(t, tmp, tt.dnsNames)
		})
	}
}
