//go:build otlp

package run

import (
	"os"
	"strings"
	"testing"
)

// TestMain skips all otel-agent run subcommand tests when running in branded mode.
// In branded builds we don't rely on the standalone otel-agent binary, so
// these tests are not relevant and can be safely skipped.
func TestMain(m *testing.M) {
	if v := os.Getenv("BRANDED"); v == "1" || strings.EqualFold(v, "true") {
		os.Exit(0)
	}
	os.Exit(m.Run())
}

