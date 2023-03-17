// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package api

import (
	"encoding/json"
	"net/http"

	"github.com/DataDog/datadog-agent/pkg/tagger"
	"github.com/DataDog/datadog-agent/pkg/tagger/collectors"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

func getTaggerList(w http.ResponseWriter, r *http.Request) {
	cardinality := collectors.TagCardinality(tagger.ChecksCardinality)
	response := tagger.List(cardinality)

	jsonTags, err := json.Marshal(response)
	if err != nil {
		setJSONError(w, log.Errorf("Unable to marshal tagger list response: %s", err), 500)
		return
	}
	w.Write(jsonTags)
}