// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018-present Datadog, Inc.

//+build !zlib

package jsonstream

import (
	"errors"
	"fmt"

	"github.com/StackVista/stackstate-agent/pkg/forwarder"
	"github.com/StackVista/stackstate-agent/pkg/serializer/marshaler"
)

const (
	// Available is true if the code is compiled in
	Available = false
)

var (
	// ErrItemTooBig is returned when a item alone exceeds maximum payload size
	ErrItemTooBig = errors.New("item alone exceeds maximum payload size")
)

// Payloads serializes a metadata payload and sends it to the forwarder
func Payloads(m marshaler.StreamJSONMarshaler) (forwarder.Payloads, error) {
	return nil, fmt.Errorf("jsonstream is not supported on this agent")
}
