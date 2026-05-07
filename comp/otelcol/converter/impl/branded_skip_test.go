//go:build otlp && test

package converterimpl

import (
	"os"
	"strings"
	"testing"
)

// TestMain skips all OTel converter tests when running in branded mode.
// In branded builds we don't rely on the embedded OTel Agent behavior tested here,
// so these tests can be skipped without affecting branded functionality.
func TestMain(m *testing.M) {
	if v := os.Getenv("BRANDED"); v == "1" || strings.EqualFold(v, "true") {
		os.Exit(0)
	}
	os.Exit(m.Run())
}
