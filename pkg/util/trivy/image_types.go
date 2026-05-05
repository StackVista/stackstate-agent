// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build trivy && (containerd || crio)

package trivy

import dockerspec "github.com/moby/docker-image-spec/specs-go/v1"

// imageInspect is a container-runtime-agnostic subset of the Docker
// image.InspectResponse type, holding only the fields the Trivy SBOM
// scanner actually reads.
type imageInspect struct {
	ID            string
	RepoTags      []string
	RepoDigests   []string
	Created       string
	Author        string
	DockerVersion string
	Architecture  string
	Os            string
	Config        *dockerspec.DockerOCIImageConfig
	RootFS        imageRootFS
}

type imageRootFS struct {
	Type   string
	Layers []string
}

type imageHistoryItem struct {
	Created   int64
	CreatedBy string
	Comment   string
	Size      int64
}
