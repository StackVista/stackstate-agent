// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

package schema

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestGenerateJSONSchemaIsInSync(t *testing.T) {
	schemaJSON, err := GenerateJSONSchema()
	require.NoError(t, err)
	moduleName := fmt.Sprintf("%s/%s", getEnv("AGENT_GITHUB_ORG", "DataDog"), getEnv("AGENT_REPO_NAME", "datadog-agent"))
	schemaWithModule := strings.ReplaceAll(string(GetDeviceProfileRcConfigJsonschema()), "DataDog/datadog-agent", moduleName)
	assert.JSONEq(t, schemaWithModule, string(schemaJSON))
}

func getEnv(key string, dfault string) string {
	value := os.Getenv(key)
	if value == "" {
		value = dfault
	}
	return value
}
