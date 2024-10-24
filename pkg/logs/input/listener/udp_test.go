// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package listener

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/StackVista/stackstate-agent/pkg/logs/config"
	"github.com/StackVista/stackstate-agent/pkg/logs/message"
	"github.com/StackVista/stackstate-agent/pkg/logs/pipeline/mock"
)

// use a randomly assigned port
var udpTestPort = 0

func TestUDPShouldReceiveMessage(t *testing.T) {
	pp := mock.NewMockProvider()
	msgChan := pp.NextPipelineChan()
	listener := NewUDPListener(pp, config.NewLogSource("", &config.LogsConfig{Port: udpTestPort}), 9000)
	listener.Start()

	conn, err := net.Dial("udp", fmt.Sprintf("%s", listener.tailer.conn.LocalAddr()))
	assert.Nil(t, err)

	var msg *message.Message

	fmt.Fprintf(conn, "hello world\n")
	msg = <-msgChan
	assert.Equal(t, "hello world", string(msg.Content))

	listener.Stop()
}
