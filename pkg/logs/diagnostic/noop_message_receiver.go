// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package diagnostic

import "github.com/StackVista/stackstate-agent/pkg/logs/message"

// NoopMessageReceiver for cases where diagnosing messages is unsupported or not needed (serverless, tests)
type NoopMessageReceiver struct{}

// HandleMessage does nothing with the message
func (n *NoopMessageReceiver) HandleMessage(m message.Message, redactedMsg []byte) {}
