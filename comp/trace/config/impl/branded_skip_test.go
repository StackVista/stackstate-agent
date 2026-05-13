// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// [sts] Mirror of comp/trace/config/branded_skip_test.go (which lives in the parent
// `package config` and so doesn't apply to tests in this `package impl` subdirectory).
// DD moved the actual trace-config tests into impl/ in 7.78.2; the existing skip mechanism
// in the parent dir no longer reaches them. STS doesn't ship the trace-agent component, so
// every test here can be safely skipped in branded builds (the env-var bindings, default
// hostnames, sites, etc. all differ from upstream after fix_branding.sh).

package configimpl

import (
	"os"
	"strings"
	"testing"
)

// TestMain skips all trace config tests when running in branded mode.
func TestMain(m *testing.M) {
	if v := os.Getenv("BRANDED"); v == "1" || strings.EqualFold(v, "true") {
		os.Exit(0)
	}
	os.Exit(m.Run())
}
