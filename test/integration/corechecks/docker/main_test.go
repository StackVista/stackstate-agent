// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package docker

import (
	"context"
	"flag"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"os"
	"strings"
	"testing"
	"time"

	log "github.com/cihub/seelog"

	"github.com/StackVista/stackstate-agent/pkg/aggregator/mocksender"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/containers/docker"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/tagger"
	"github.com/StackVista/stackstate-agent/pkg/tagger/collectors"
	"github.com/StackVista/stackstate-agent/pkg/tagger/local"
	"github.com/StackVista/stackstate-agent/pkg/workloadmeta"
	"github.com/StackVista/stackstate-agent/test/integration/utils"

	_ "github.com/StackVista/stackstate-agent/pkg/workloadmeta/collectors"
)

var (
	retryDelay   = flag.Int("retry-delay", 1, "time to wait between retries (default 1 second)")
	retryTimeout = flag.Int("retry-timeout", 30, "maximum time before failure (default 30 seconds)")
	skipCleanup  = flag.Bool("skip-cleanup", false, "skip cleanup of the docker containers (for debugging)")
)

var dockerCfgString = `
collect_events: true
collect_container_size: true
collect_images_stats: true
collect_exit_codes: true
tags:
  - instanceTag:MustBeHere
`

var datadogCfgString = `
docker_labels_as_tags:
    "high.card.label": +highcardlabeltag
    "low.card.label": lowcardlabeltag
docker_env_as_tags:
    "HIGH_CARD_ENV": +highcardenvtag
    "low_card_env": lowcardenvtag
`

var (
	sender      *mocksender.MockSender
	dockerCheck check.Check
)

func TestMain(m *testing.M) {
	flag.Parse()

	config.SetupLogger(
		config.LoggerName("test"),
		"debug",
		"",
		"",
		false,
		true,
		false,
	)

	retryTicker := time.NewTicker(time.Duration(*retryDelay) * time.Second)
	timeoutTicker := time.NewTicker(time.Duration(*retryTimeout) * time.Second)
	var lastRunResult int
	var retryCount int

	err := setup()
	if err != nil {
		log.Infof("Test setup failed: %v", err)
		tearOffAndExit(1)
	}

	for {
		select {
		case <-retryTicker.C:
			retryCount++
			log.Infof("Starting run %d", retryCount)
			lastRunResult = doRun(m)
			if lastRunResult == 0 {
				tearOffAndExit(0)
			}
		case <-timeoutTicker.C:
			log.Errorf("Timeout after %d seconds and %d retries", retryTimeout, retryCount)
			tearOffAndExit(1)
		}
	}
}

// Called before for first test run: compose up
func setup() error {
	// Setup global conf
	config.Datadog.SetConfigType("yaml")
	err := config.Datadog.ReadConfig(strings.NewReader(datadogCfgString))
	if err != nil {
		return err
	}
	config.DetectFeatures()

	workloadmeta.GetGlobalStore().Start(context.Background())

	// Setup tagger
	tagger.SetDefaultTagger(local.NewTagger(collectors.DefaultCatalog))
	tagger.Init()

	// Start compose recipes
	for projectName, file := range defaultCatalog.composeFilesByProjects {
		compose := &utils.ComposeConf{
			ProjectName: projectName,
			FilePath:    file,
		}
		output, err := compose.Start()
		if err != nil {
			log.Errorf("Compose didn't start properly: %s", string(output))
			return err
		}
	}
	return nil
}

// Reset the state and trigger a new run
func doRun(m *testing.M) int {
	// Setup docker check
	dockerCfg := []byte(dockerCfgString)
	dockerInitCfg := []byte("")
	dockerCheck = docker.DockerFactory()
	dockerCheck.Configure(dockerCfg, dockerInitCfg, "test")

	// Setup mock sender
	sender = mocksender.NewMockSender(dockerCheck.ID())
	sender.SetupAcceptAll()

	// Setup mock batcher
	_ = batcher.NewMockBatcher()

	err := dockerCheck.Run()
	if err != nil {
		log.Errorf("Docker check run error: %s", err)
	}
	return m.Run()
}

// Compose down, reset the external states and exit
func tearOffAndExit(exitcode int) {
	if *skipCleanup {
		os.Exit(exitcode)
	}

	// Stop compose recipes, ignore errors
	for projectName, file := range defaultCatalog.composeFilesByProjects {
		compose := &utils.ComposeConf{
			ProjectName: projectName,
			FilePath:    file,
		}
		output, err := compose.Stop()
		if err != nil {
			log.Warnf("Compose didn't stop properly: %s", string(output))
		}
	}
	os.Exit(exitcode)
}
