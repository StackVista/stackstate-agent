// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build jmx
// +build jmx

package jmx

import (
	"fmt"
	"time"

	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	telemetry_utils "github.com/StackVista/stackstate-agent/pkg/telemetry/utils"
	"github.com/StackVista/stackstate-agent/pkg/util/features"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

type JMXCheck struct {
	id        check.ID
	name      string
	config    integration.Config
	stop      chan struct{}
	source    string
	telemetry bool
	features  features.Features
}

func newJMXCheck(config integration.Config, source string) *JMXCheck {
	check := &JMXCheck{
		config:    config,
		stop:      make(chan struct{}),
		name:      config.Name,
		id:        check.ID(fmt.Sprintf("%v_%v", config.Name, config.Digest())),
		source:    source,
		telemetry: telemetry_utils.IsCheckEnabled("jmx"),
	}
	check.Configure(config.InitConfig, config.MetricConfig, source) //nolint:errcheck

	return check
}

func (c *JMXCheck) Run() error {
	err := state.scheduleCheck(c)
	if err != nil {
		return err
	}

	select {
	case <-state.runnerError:
		return fmt.Errorf("jmxfetch exited, stopping %s : %s", c.name, err)
	case <-c.stop:
		log.Infof("jmx check %s stopped", c.name)
	}

	return nil
}

func (c *JMXCheck) Stop() {
	close(c.stop)
	state.unscheduleCheck(c)
}

func (c *JMXCheck) Cancel() {}

func (c *JMXCheck) String() string {
	return c.name
}

func (c *JMXCheck) Version() string {
	return ""
}

func (c *JMXCheck) ConfigSource() string {
	return c.source
}

func (c *JMXCheck) Configure(config integration.Data, initConfig integration.Data, source string) error {
	return nil
}

func (c *JMXCheck) Interval() time.Duration {
	return 0
}

func (c *JMXCheck) ID() check.ID {
	return c.id
}

func (c *JMXCheck) IsTelemetryEnabled() bool {
	return c.telemetry
}

func (c *JMXCheck) GetWarnings() []error {
	return []error{}
}

func (c *JMXCheck) GetSenderStats() (check.SenderStats, error) {
	return check.NewSenderStats(), nil
}

// GetFeatures returns the features supported by StackState
func (c *JMXCheck) GetFeatures() features.Features {
	return c.features
}

// SetFeatures sets the features supported by StackState
func (c *JMXCheck) SetFeatures(features features.Features) {
	c.features = features
}
