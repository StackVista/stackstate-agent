package container_runtime

import (
	"github.com/opencontainers/runtime-spec/specs-go"
)

type Container struct {
	Name   string
	Type   string
	ID     string
	Image  string
	Mounts []specs.Mount
	State  string
	Health string
}

type ContainerUtil interface {
	GetContainers() ([]*Container, error)
}
