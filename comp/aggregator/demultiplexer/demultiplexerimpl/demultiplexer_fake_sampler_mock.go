// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

//go:build test

package demultiplexerimpl

import (
	"time"

	demultiplexerComp "github.com/DataDog/datadog-agent/comp/aggregator/demultiplexer"
	"github.com/DataDog/datadog-agent/comp/core/log"
	"github.com/DataDog/datadog-agent/pkg/aggregator"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"go.uber.org/fx"
)

// FakeSamplerMockModule defines the fx options for FakeSamplerMock.
func FakeSamplerMockModule() fxutil.Module {
	return fxutil.Component(
		fx.Provide(newFakeSamplerMock))
}

type fakeSamplerMockDependencies struct {
	fx.In
	Log log.Component
}

type fakeSamplerMock struct {
	*TestAgentDemultiplexer
}

func (f fakeSamplerMock) GetAgentDemultiplexer() *aggregator.AgentDemultiplexer {
	return f.TestAgentDemultiplexer.AgentDemultiplexer
}

func newFakeSamplerMock(deps fakeSamplerMockDependencies) demultiplexerComp.FakeSamplerMock {
	demux := initTestAgentDemultiplexerWithFlushInterval(deps.Log, time.Hour)
	return fakeSamplerMock{
		TestAgentDemultiplexer: demux,
	}
}