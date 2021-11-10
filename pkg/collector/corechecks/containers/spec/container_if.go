package spec

import (
	"github.com/opencontainers/runtime-spec/specs-go"
)

type Container struct {
	Name    string        `json:"name"`
	Runtime string        `json:"runtime"`
	ID      string        `json:"id"`
	Image   string        `json:"image,omitempty"`
	Mounts  []specs.Mount `json:"mounts"`
	State   string        `json:"state,omitempty"`
	Tags    []string      `json:"tags"`
}

type ContainerUtil interface {
	GetContainers() ([]*Container, error)
}
