package transactional

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/batcher"
	"github.com/StackVista/stackstate-agent/pkg/collector/check"
	"time"
)

// MakeCheckInstanceBatcher initializes the batcher instance for a check instance
func MakeCheckInstanceBatcher(checkId check.ID, hostname, agentName string, maxCapacity int, flushInterval time.Duration) *CheckTransactionalBatcher {
	checkFlushInterval := time.NewTicker(flushInterval)
	ctb := &CheckTransactionalBatcher{
		BatcherBase:   batcher.MakeBatcherBase(hostname, agentName, maxCapacity),
		CheckInstance: checkId,
		flushTicker:   checkFlushInterval,
	}

	go ctb.listenForFlushTicker()

	return ctb
}

// CheckTransactionalBatcher is a instance of a batcher for a specific check instance
type CheckTransactionalBatcher struct {
	batcher.BatcherBase
	CheckInstance check.ID
	builder       TopologyBuilder
	flushTicker   *time.Ticker
}

// GetCheckInstance returns the check instance for this batcher
func (ctb *CheckTransactionalBatcher) GetCheckInstance() check.ID {
	return ctb.CheckInstance
}

// FlushCheckInstance submits complete for this check instance
func (ctb *CheckTransactionalBatcher) FlushCheckInstance() {
	ctb.SubmitComplete(ctb.CheckInstance)
}

// listenForFlushTicker waits for messages on the ticker channel and submits a flush for this check
func (ctb *CheckTransactionalBatcher) listenForFlushTicker() {
	for _ = range ctb.flushTicker.C {
		ctb.FlushCheckInstance()
	}
}

func (ctb *CheckTransactionalBatcher) Start() {
	for {
		s := <-ctb.Input
		switch submission := s.(type) {
		case batcher.SubmitComponent:
			ctb.builder.AddComponent(submission.CheckID, submission.Instance, submission.Component)
		case batcher.SubmitRelation:
			ctb.builder.AddRelation(submission.CheckID, submission.Instance, submission.Relation)
		case batcher.SubmitStartSnapshot:
			ctb.builder.StartSnapshot(submission.CheckID, submission.Instance)
		case batcher.SubmitStopSnapshot:
			ctb.builder.StopSnapshot(submission.CheckID, submission.Instance)
		case batcher.SubmitHealthCheckData:

		case batcher.SubmitHealthStartSnapshot:
		case batcher.SubmitHealthStopSnapshot:
		case batcher.SubmitRawMetricsData:
		case batcher.SubmitComplete:
		case batcher.SubmitShutdown:
			return
		default:
			panic(fmt.Sprint("Unknown submission type"))
		}
	}
}
