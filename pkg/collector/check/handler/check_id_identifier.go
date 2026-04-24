package handler

import checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"

// NewCheckIdentifier returns a IDOnlyCheckIdentifier that implements the CheckIdentifier interface for a given check ID.
func NewCheckIdentifier(checkID checkid.ID) CheckIdentifier {
	return &IDOnlyCheckIdentifier{checkID: checkID}
}

// IDOnlyCheckIdentifier is used to create a CheckIdentifier when only the checkID is present
type IDOnlyCheckIdentifier struct {
	checkID checkid.ID
}

// String returns the IDOnlyCheckIdentifier checkID as a string
func (idCI *IDOnlyCheckIdentifier) String() string {
	return string(idCI.checkID)
}

// ID returns the IDOnlyCheckIdentifier checkID
func (idCI *IDOnlyCheckIdentifier) ID() checkid.ID {
	return idCI.checkID
}

// ConfigSource returns an empty string for the IDOnlyCheckIdentifier
func (*IDOnlyCheckIdentifier) ConfigSource() string {
	return ""
}
