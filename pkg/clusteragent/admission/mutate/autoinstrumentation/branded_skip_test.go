//go:build kubeapiserver

package autoinstrumentation

import (
	"os"
	"strings"
	"testing"
)

// TestMain skips all autoinstrumentation config tests when running in branded mode.
// In branded builds we don't rely on APM instrumentation features, so
// these tests are not relevant and can be safely skipped.
func TestMain(m *testing.M) {
	if v := os.Getenv("BRANDED"); v == "1" || strings.EqualFold(v, "true") {
		os.Exit(0)
	}
	os.Exit(m.Run())
}
