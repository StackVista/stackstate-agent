// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build cri
// +build cri

package cri

import (
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/containers/topology"
	"github.com/StackVista/stackstate-agent/pkg/metrics"
	"time"

	yaml "gopkg.in/yaml.v2"
	pb "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"

	"github.com/StackVista/stackstate-agent/pkg/aggregator"
	"github.com/StackVista/stackstate-agent/pkg/autodiscovery/integration"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	core "github.com/StackVista/stackstate-agent/pkg/collector/corechecks"
	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/tagger"
	"github.com/StackVista/stackstate-agent/pkg/tagger/collectors"
	"github.com/StackVista/stackstate-agent/pkg/util/containers"
	"github.com/StackVista/stackstate-agent/pkg/util/containers/cri"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

const (
	criCheckName = "cri"
)

// CRIConfig holds the config of the check
type CRIConfig struct {
	CollectDisk bool `yaml:"collect_disk"`
	// sts
	CollectContainerTopology bool `yaml:"collect_container_topology"`
}

// CRICheck grabs CRI metrics
type CRICheck struct {
	core.CheckBase
	instance *CRIConfig
	filter   *containers.Filter
	// sts
	topologyCollector *topology.ContainerTopologyCollector
}

func init() {
	core.RegisterCheck("cri", CRIFactory)
}

// CRIFactory is exported for integration testing
func CRIFactory() check.Check {
	return &CRICheck{
		CheckBase: core.NewCheckBase(criCheckName),
		instance:  &CRIConfig{},
		// sts
		topologyCollector: topology.MakeContainerTopologyCollector(criCheckName),
	}
}

// Parse parses the CRICheck config and set default values
func (c *CRIConfig) Parse(data []byte) error {
	// default values
	c.CollectDisk = false
	// sts
	c.CollectContainerTopology = true

	return yaml.Unmarshal(data, c)
}

// Configure parses the check configuration and init the check
func (c *CRICheck) Configure(config, initConfig integration.Data, source string) error {
	err := c.CommonConfigure(config, source)
	if err != nil {
		return err
	}

	filter, err := containers.GetSharedMetricFilter()
	if err != nil {
		return err
	}
	c.filter = filter

	return c.instance.Parse(config)
}

// Run executes the check
func (c *CRICheck) Run() error {
	sender, err := aggregator.GetSender(c.ID())
	if err != nil {
		return err
	}

	util, err := cri.GetUtil()
	if err != nil {
		c.Warnf("Error initialising check: %s", err) //nolint:errcheck
		return err
	}

	containerStats, err := util.ListContainerStats()
	if err != nil {
		c.Warnf("Cannot get containers from the CRI: %s", err) //nolint:errcheck
		return err
	}
	c.generateMetrics(sender, containerStats, util)

	// sts begin
	// Collect container topology
	if c.instance.CollectContainerTopology && !config.IsFeaturePresent(config.Containerd) {
		err := c.topologyCollector.BuildContainerTopology(util)
		if err != nil {
			sender.ServiceCheck("cri.health", metrics.ServiceCheckCritical, "", nil, err.Error())
			log.Errorf("Could not collect container topology: %s", err)
			return err
		}
	}
	//sts end

	sender.Commit()
	return nil
}

func (c *CRICheck) generateMetrics(sender aggregator.Sender, containerStats map[string]*pb.ContainerStats, criUtil cri.CRIClient) {
	for cid, stats := range containerStats {
		if stats == nil {
			log.Warnf("Missing stats for container: %s", cid)
			continue
		}

		ctnStatus, err := criUtil.GetContainerStatus(cid)
		if err != nil {
			log.Debugf("Could not retrieve the status of container %q: %s", cid, err)
			continue
		}

		if ctnStatus == nil || ctnStatus.State != pb.ContainerState_CONTAINER_RUNNING || c.isExcluded(ctnStatus) {
			continue
		}

		entityID := containers.BuildTaggerEntityName(cid)
		tags, err := tagger.Tag(entityID, collectors.HighCardinality)
		if err != nil {
			log.Errorf("Could not collect tags for container %s: %s", cid[:12], err)
		}
		tags = append(tags, "runtime:"+criUtil.GetRuntime())

		currentUnixTime := time.Now().UnixNano()
		c.computeContainerUptime(sender, currentUnixTime, *ctnStatus, tags)

		c.processContainerStats(sender, *stats, tags)
	}
}

// processContainerStats extracts metrics from the protobuf object
func (c *CRICheck) processContainerStats(sender aggregator.Sender, stats pb.ContainerStats, tags []string) {
	sender.Gauge("cri.mem.rss", float64(stats.GetMemory().GetWorkingSetBytes().GetValue()), "", tags)
	// Cumulative CPU usage (sum across all cores) since object creation.
	sender.Rate("cri.cpu.usage", float64(stats.GetCpu().GetUsageCoreNanoSeconds().GetValue()), "", tags)
	if c.instance.CollectDisk {
		sender.Gauge("cri.disk.used", float64(stats.GetWritableLayer().GetUsedBytes().GetValue()), "", tags)
		sender.Gauge("cri.disk.inodes", float64(stats.GetWritableLayer().GetInodesUsed().GetValue()), "", tags)
	}
}

func (c *CRICheck) computeContainerUptime(sender aggregator.Sender, currentTime int64, ctnStatus pb.ContainerStatus, tags []string) {
	if ctnStatus.StartedAt != 0 && currentTime-ctnStatus.StartedAt > 0 {
		sender.Gauge("cri.uptime", float64((currentTime-ctnStatus.StartedAt)/int64(time.Second)), "", tags)
	}
}

// isExcluded returns whether a container should be excluded based on its image, name and namespace
func (c *CRICheck) isExcluded(ctr *pb.ContainerStatus) bool {
	if config.Datadog.GetBool("exclude_pause_container") && containers.IsPauseContainer(ctr.Labels) {
		return true
	}

	name := ""
	if meta := ctr.GetMetadata(); meta != nil {
		name = meta.GetName()
	}

	image := ""
	if imSpec := ctr.GetImage(); imSpec != nil {
		image = imSpec.GetImage()
	}

	return c.filter.IsExcluded(name, image, ctr.GetLabels()["io.kubernetes.pod.namespace"])
}