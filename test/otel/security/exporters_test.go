// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2026-present Datadog, Inc.

package security_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	logpb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Exercise the real exporters with verbose diagnostics, where CVE-2026-81870
// exposed endpoint configuration during provider construction.
func TestTraceDiagnosticsOmitEndpoint(t *testing.T) {
	const endpoint = "private-collector.example:4317"
	for _, protocol := range []string{"grpc", "http"} {
		t.Run(protocol, func(t *testing.T) {
			var output bytes.Buffer
			otel.SetLogger(funcr.New(func(_, message string) { output.WriteString(message) }, funcr.Options{Verbosity: 4}))
			t.Cleanup(func() { otel.SetLogger(logr.Discard()) })
			var exporter sdktrace.SpanExporter
			var err error
			if protocol == "grpc" {
				exporter, err = otlptracegrpc.New(t.Context(), otlptracegrpc.WithEndpoint(endpoint))
			} else {
				exporter, err = otlptracehttp.New(t.Context(), otlptracehttp.WithEndpoint(endpoint))
			}
			require.NoError(t, err)
			provider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
			require.NoError(t, provider.Shutdown(t.Context()))
			require.Contains(t, output.String(), "TracerProvider created")
			require.NotContains(t, output.String(), endpoint)
		})
	}
}

type logCollector struct {
	logpb.UnimplementedLogsServiceServer
	received chan struct{}
}

func (c *logCollector) Export(context.Context, *logpb.ExportLogsServiceRequest) (*logpb.ExportLogsServiceResponse, error) {
	c.received <- struct{}{}
	return &logpb.ExportLogsServiceResponse{}, nil
}

// CVE-2026-81871 ignored environment TLS settings. A private CA and required
// client authentication ensure this cannot pass using system roots alone.
func TestLogExporterEnvironmentMTLS(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "OTel regression"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true,
		KeyUsage:    x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)
	roots := x509.NewCertPool()
	require.True(t, roots.AppendCertsFromPEM(certPEM))
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	server := grpc.NewServer(grpc.Creds(credentials.NewTLS(&tls.Config{
		MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{cert},
		ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: roots,
	})))
	collector := &logCollector{received: make(chan struct{}, 2)}
	logpb.RegisterLogsServiceServer(server, collector)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	dir := t.TempDir()
	certPath, keyPath := filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
	require.NoError(t, os.WriteFile(certPath, certPEM, 0600))
	require.NoError(t, os.WriteFile(keyPath, keyPEM, 0600))
	// Neutralize inherited exporter options so only the test's trust material applies.
	for _, variable := range os.Environ() {
		name, _, _ := strings.Cut(variable, "=")
		if strings.HasPrefix(name, "OTEL_EXPORTER_OTLP") {
			t.Setenv(name, "")
		}
	}
	for _, prefix := range []string{"OTEL_EXPORTER_OTLP", "OTEL_EXPORTER_OTLP_LOGS"} {
		t.Run(prefix, func(t *testing.T) {
			t.Setenv(prefix+"_ENDPOINT", "https://"+listener.Addr().String())
			t.Setenv(prefix+"_CERTIFICATE", certPath)
			t.Setenv(prefix+"_CLIENT_CERTIFICATE", certPath)
			t.Setenv(prefix+"_CLIENT_KEY", keyPath)
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			defer cancel()
			exporter, err := otlploggrpc.New(ctx)
			require.NoError(t, err)
			defer func() { require.NoError(t, exporter.Shutdown(context.Background())) }()
			require.NoError(t, exporter.Export(ctx, []sdklog.Record{{}}))
			select {
			case <-collector.received:
			case <-ctx.Done():
				t.Fatal("collector did not receive the log")
			}
		})
	}
}
