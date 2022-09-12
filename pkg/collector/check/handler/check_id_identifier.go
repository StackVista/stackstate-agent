package handler

import "github.com/StackVista/stackstate-agent/pkg/collector/check"

// NewCheckIdentifier returns a IDOnlyCheckIdentifier that implements the CheckIdentifier interface for a given check ID.
func NewCheckIdentifier(checkID check.ID) CheckIdentifier {
	return &IDOnlyCheckIdentifier{checkID: checkID}
}

// IDOnlyCheckIdentifier is used to create a CheckIdentifier when only the checkID is present
type IDOnlyCheckIdentifier struct {
	checkID check.ID
}

// String returns the IDOnlyCheckIdentifier checkID as a string
func (idCI *IDOnlyCheckIdentifier) String() string {
	return string(idCI.checkID)
}

// ID returns the IDOnlyCheckIdentifier checkID
func (idCI *IDOnlyCheckIdentifier) ID() check.ID {
	return idCI.checkID
}

// ConfigSource returns an empty string for the IDOnlyCheckIdentifier
func (*IDOnlyCheckIdentifier) ConfigSource() string {
	return ""
}
