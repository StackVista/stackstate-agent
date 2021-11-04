package util

import "github.com/docker/docker/api/types"

type Container struct {
	Name   string
	Type   string
	ID     string
	Image  string
	Mounts []types.MountPoint
	State  string
	Health string
}

type ContainerUtil interface {
	GetContainers() ([]*Container, error)
}
