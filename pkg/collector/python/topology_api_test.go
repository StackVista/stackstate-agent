//go:build python && test
// +build python,test

package python

import "testing"

func TestComponentTopology(t *testing.T) {
	testComponentTopology(t)
}

func TestRelationTopology(t *testing.T) {
	testRelationTopology(t)
}

func TestStartTransaction(t *testing.T) {
	testStartTransaction(t)
}

func TestStopTransaction(t *testing.T) {
	testStopTransaction(t)
}

func TestStartSnapshotCheck(t *testing.T) {
	testStartSnapshotCheck(t)
}

func TestStopSnapshotCheck(t *testing.T) {
	testStopSnapshotCheck(t)
}

func TestDeleteTopologyElement(t *testing.T) {
	testDeleteTopologyElement(t)
}
