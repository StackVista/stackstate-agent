// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build cri
// +build cri

package cri

import (
	"context"
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/collector/corechecks/containers/spec"
	"github.com/StackVista/stackstate-agent/pkg/util/containers"
	"github.com/opencontainers/runtime-spec/specs-go"
	"net"
	"sync"
	"time"

	"github.com/StackVista/stackstate-agent/pkg/config"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"github.com/StackVista/stackstate-agent/pkg/util/retry"
	"google.golang.org/grpc"
	pb "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

// sts begin
const (
	// Default CRI socket path
	criDefaultSocketPath = "/var/run/crio/crio.sock"
)

// sts end

var (
	globalCRIUtil *CRIUtil
	once          sync.Once
)

type CRIClient interface {
	ListContainerStats() (map[string]*pb.ContainerStats, error)
	GetContainerStatus(containerID string) (*pb.ContainerStatus, error)
	GetRuntime() string
	GetRuntimeVersion() string
}

// CRIUtil wraps interactions with the CRI and implements CRIClient
// see https://github.com/kubernetes/kubernetes/blob/release-1.12/pkg/kubelet/apis/cri/runtime/v1alpha2/api.proto
type CRIUtil struct {
	// used to setup the CRIUtil
	initRetry retry.Retrier

	sync.Mutex
	client            pb.RuntimeServiceClient
	runtime           string
	runtimeVersion    string
	queryTimeout      time.Duration
	connectionTimeout time.Duration
	socketPath        string
}

// init makes an empty CRIUtil bootstrap itself.
// This is not exposed as public API but is called by the retrier embed.
func (c *CRIUtil) init() error {
	if c.socketPath == "" {
		return fmt.Errorf("no cri_socket_path was set")
	}

	dialer := func(socketPath string, timeout time.Duration) (net.Conn, error) {
		return net.DialTimeout("unix", socketPath, timeout)
	}

	conn, err := grpc.Dial(c.socketPath, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(c.connectionTimeout), grpc.WithDialer(dialer))
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}

	c.client = pb.NewRuntimeServiceClient(conn)
	// validating the connection by fetching the version
	ctx, cancel := context.WithTimeout(context.Background(), c.connectionTimeout)
	defer cancel()
	request := &pb.VersionRequest{}
	r, err := c.client.Version(ctx, request)
	if err != nil {
		return err
	}
	c.runtime = r.RuntimeName
	c.runtimeVersion = r.RuntimeVersion
	log.Debugf("Successfully connected to CRI %s %s", c.runtime, c.runtimeVersion)

	return nil
}

// GetUtil returns a ready to use CRIUtil. It is backed by a shared singleton.
func GetUtil() (*CRIUtil, error) {
	once.Do(func() {
		globalCRIUtil = &CRIUtil{
			queryTimeout:      config.Datadog.GetDuration("cri_query_timeout") * time.Second,
			connectionTimeout: config.Datadog.GetDuration("cri_connection_timeout") * time.Second,
			socketPath:        config.Datadog.GetString("cri_socket_path"),
		}
		// sts begin
		if globalCRIUtil.socketPath == "" {
			log.Info("No socket path was specified, defaulting to /var/run/crio/crio.sock")
			globalCRIUtil.socketPath = criDefaultSocketPath
		}
		// sts end
		globalCRIUtil.initRetry.SetupRetrier(&retry.Config{ //nolint:errcheck
			Name:              "criutil",
			AttemptMethod:     globalCRIUtil.init,
			Strategy:          retry.Backoff,
			InitialRetryDelay: 1 * time.Second,
			MaxRetryDelay:     5 * time.Minute,
		})
	})

	if err := globalCRIUtil.initRetry.TriggerRetry(); err != nil {
		log.Debugf("CRI init error: %s", err)
		return nil, err
	}
	return globalCRIUtil, nil
}

// sts begin

// ContainerStateMap is used to map cri specific state to own internal state
var ContainerStateMap = map[pb.ContainerState]string{
	pb.ContainerState_CONTAINER_CREATED: containers.ContainerCreatedState,
	pb.ContainerState_CONTAINER_RUNNING: containers.ContainerRunningState,
	pb.ContainerState_CONTAINER_EXITED:  containers.ContainerExitedState,
	pb.ContainerState_CONTAINER_UNKNOWN: containers.ContainerUnknownState,
}

func (c *CRIUtil) GetContainers(ctx context.Context) ([]*spec.Container, error) {
	containerStats, err := c.ListContainerStats()
	if err != nil {
		return nil, err
	}
	uContainers := make([]*spec.Container, 0, len(containerStats))
	for cid := range containerStats {
		cstatus, err := c.GetContainerStatus(cid)
		if err != nil {
			_ = log.Warnf("Could not get Status from CRI container '%s'. Error: %v", cid, err)
			continue
		}
		if cstatus.Metadata == nil {
			_ = log.Warnf("Could not get Metadata from CRI container '%s'", cid)
			continue
		}
		if cstatus.Image == nil {
			_ = log.Warnf("Could not get Image from CRI container '%s'", cid)
			continue
		}
		mounts := make([]specs.Mount, 0, len(cstatus.Mounts))
		for _, cmount := range cstatus.Mounts {
			mountPoint := specs.Mount{
				Source:      cmount.HostPath,
				Destination: cmount.ContainerPath,
			}
			mounts = append(mounts, mountPoint)
		}
		container := &spec.Container{
			Runtime: "cri",
			Name:    cstatus.Metadata.Name,
			ID:      cid,
			Image:   cstatus.Image.Image,
			Mounts:  mounts,
		}
		if state, ok := ContainerStateMap[cstatus.State]; ok {
			container.State = state
		} else {
			_ = log.Warnf("Could not map state of container '%s'. State: %s", cid, cstatus.State.String())
		}
		uContainers = append(uContainers, container)
	}
	return uContainers, nil
}

// sts end

// ListContainerStats sends a ListContainerStatsRequest to the server, and parses the returned response
func (c *CRIUtil) ListContainerStats() (map[string]*pb.ContainerStats, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.queryTimeout)
	defer cancel()
	filter := &pb.ContainerStatsFilter{}
	request := &pb.ListContainerStatsRequest{Filter: filter}
	r, err := c.client.ListContainerStats(ctx, request)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]*pb.ContainerStats)
	for _, s := range r.GetStats() {
		stats[s.Attributes.Id] = s
	}
	return stats, nil
}

// GetContainerStatus requests a container status by its ID
func (c *CRIUtil) GetContainerStatus(containerID string) (*pb.ContainerStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.queryTimeout)
	defer cancel()
	request := &pb.ContainerStatusRequest{ContainerId: containerID}
	r, err := c.client.ContainerStatus(ctx, request)
	if err != nil {
		return nil, err
	}

	return r.Status, nil
}

func (c *CRIUtil) GetRuntime() string {
	return c.runtime
}

func (c *CRIUtil) GetRuntimeVersion() string {
	return c.runtimeVersion
}
