// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build docker

package compliance

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	docker "github.com/docker/docker/client"
)

func (r *defaultResolver) resolveDocker(ctx context.Context, spec InputSpecDocker) (interface{}, error) {
	cl, ok := r.dockerCl.(docker.APIClient)
	if !ok || cl == nil {
		return nil, ErrIncompatibleEnvironment
	}

	var resolved []interface{}
	switch spec.Kind {
	case "image":
		list, err := cl.ImageList(ctx, image.ListOptions{All: true})
		if err != nil {
			return nil, err
		}
		for _, im := range list {
			image, err := cl.ImageInspect(ctx, im.ID)
			if err != nil {
				return nil, err
			}
			resolved = append(resolved, map[string]interface{}{
				"id":      image.ID,
				"tags":    image.RepoTags,
				"inspect": image,
			})
		}
	case "container":
		list, err := cl.ContainerList(ctx, container.ListOptions{All: true})
		if err != nil {
			return nil, err
		}
		for _, cn := range list {
			container, _, err := cl.ContainerInspectWithRaw(ctx, cn.ID, false)
			if err != nil {
				return nil, err
			}
			imageRepo := parseImageRepo(container.Config.Image)
			resolved = append(resolved, map[string]interface{}{
				"id":         container.ID,
				"name":       container.Name,
				"image":      container.Image,
				"image_repo": imageRepo,
				"inspect":    container,
			})
		}
	case "network":
		networks, err := cl.NetworkList(ctx, network.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, nw := range networks {
			resolved = append(resolved, map[string]interface{}{
				"id":      nw.ID,
				"name":    nw.Name,
				"inspect": nw,
			})
		}
	case "info":
		info, err := cl.Info(ctx)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, map[string]interface{}{
			"inspect": info,
		})
	case "version":
		version, err := cl.ServerVersion(ctx)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, map[string]interface{}{
			"version":       version.Version,
			"apiVersion":    version.APIVersion,
			"platform":      version.Platform.Name,
			"experimental":  version.Experimental,
			"os":            version.Os,
			"arch":          version.Arch,
			"kernelVersion": version.KernelVersion,
		})
	default:
		return nil, fmt.Errorf("unsupported docker object kind '%q'", spec.Kind)
	}

	return resolved, nil
}
