package spec

import (
	"github.com/opencontainers/runtime-spec/specs-go"
)

// Container holds the data to be sent when creating a container component
type Container struct {
	Name    string        `json:"name"`
	Runtime string        `json:"runtime"`
	ID      string        `json:"id"`
	Image   string        `json:"image,omitempty"`
	Mounts  []specs.Mount `json:"mounts"`
	State   string        `json:"state,omitempty"`
}

// ContainerUtil is an interface for util classes capable of getting a list of Container
type ContainerUtil interface {
	GetContainers() ([]*Container, error)
}
