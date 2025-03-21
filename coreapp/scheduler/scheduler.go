package scheduler

import (
	"fmt"
	"sync"

	"github.com/oqtopus-team/oqtopus-engine/coreapp/core"
	"go.uber.org/zap"
)

type statusHistory map[string][]core.Status

type NormalScheduler struct {
	queue         *NormalQueue
	statusHistory statusHistory
	// TODO: lock
}

type jobInScheduler struct {
	job      core.Job
	finished *sync.WaitGroup
}

func (n *NormalScheduler) Setup(conf *core.Conf) error {
	n.queue = &NormalQueue{}
	n.queue.Setup(conf)
	n.statusHistory = make(statusHistory)
	return nil
}

// TODO: use rungroup
func (n *NormalScheduler) Start() error {
	// TODO: functionalize
	go func() {
		for {
			zap.L().Debug("checking the queue...")
			jis, err := n.queue.Dequeue(true)
			jid := jis.job.JobData().ID
			if err != nil {
				zap.L().Error(fmt.Sprintf("failed to get job(%s) from queue. Reason:%s", jid, err))
				continue
			}
			zap.L().Debug(fmt.Sprintf("processing job:%s", jid))
			// TODO: not update status in scheduler
			st := core.RUNNING
			n.statusHistory[jid] = append(n.statusHistory[jid], st)
			jis.job.JobData().Status = st
			jis.job.JobContext().DBChan <- jis.job.Clone()
			jis.job.Process()
			zap.L().Debug(fmt.Sprintf("finished to process job(%s), status:%s", jid, jis.job.JobData().Status))
			jis.finished.Done()
		}
	}()
	// TODO connected Channel
	return nil
}

func (n *NormalScheduler) HandleJob(j core.Job) {
	zap.L().Debug(fmt.Sprintf("starting to handle job(%s) in %s", j.JobData().ID, j.JobData().Status))
	go func() {
		defer func() {
			zap.L().Debug(fmt.Sprintf("status history job(%s): %v", j.JobData().ID, n.statusHistory[j.JobData().ID]))
			delete(n.statusHistory, j.JobData().ID)
		}()
		n.handleImpl(j)
	}()
}

func (n *NormalScheduler) HandleJobForTest(j core.Job, wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()
		n.handleImpl(j)
	}()
}

func (n *NormalScheduler) handleImpl(j core.Job) {
	for {
		jid := j.JobData().ID
		j.JobData().UseJobInfoUpdate = false //very adhoc
		st := j.JobData().Status             // must be ready
		n.statusHistory[jid] = append(n.statusHistory[jid], st)
		zap.L().Debug(fmt.Sprintf("handling job(%s)in %s starting", jid, st))
		if j.JobData().Status != core.READY {
			zap.L().Error(
				fmt.Sprintf("finished to handle job(%s) with unexpected status:%s", jid, j.JobData().Status.String()))
			// not write to DB
			return
		}
		zap.L().Debug(fmt.Sprintf("handling job(%s). start pre-processing", jid))
		j.PreProcess()
		if j.IsFinished() {
			j.JobData().UseJobInfoUpdate = true //TODO: fix this adhoc
		}
		j.JobContext().DBChan <- j.Clone()
		if j.IsFinished() {
			zap.L().Debug(fmt.Sprintf("finished to handle job(%s) after pre-processing", jid))
			n.statusHistory[jid] = append(n.statusHistory[jid], j.JobData().Status)
			return
		}
		var wg sync.WaitGroup
		wg.Add(1)
		jis := &jobInScheduler{
			job:      j,
			finished: &wg,
		}
		n.queue.queueChan <- jis
		wg.Wait()                           // wait for processing
		j.JobData().UseJobInfoUpdate = true //TODO: fix this adhoc
		zap.L().Debug(fmt.Sprintf("Processed Job Status: %s", j.JobData().Status))
		if j.IsFinished() {
			j.JobContext().DBChan <- j.Clone()
			zap.L().Debug(fmt.Sprintf("finished to handle job(%s) after processing with status:%s",
				jid, j.JobData().Status.String()))
			n.statusHistory[jid] = append(n.statusHistory[jid], j.JobData().Status)
			j.JobContext().DBChan <- j.Clone()
			return
		}
		zap.L().Debug(fmt.Sprintf("handling job(%s). start post-processing", jid))
		j.PostProcess()
		if j.IsFinished() {
			zap.L().Debug(fmt.Sprintf("finished to handle job(%s) after post-processing with status:%s",
				jid, j.JobData().Status.String()))
			n.statusHistory[jid] = append(n.statusHistory[jid], j.JobData().Status)
			j.JobContext().DBChan <- j.Clone()
			return
		}
		zap.L().Debug(fmt.Sprintf("one more loop for job(%s)", jid))
	}
}

func (n *NormalScheduler) GetCurrentQueueSize() int {
	return n.queue.fifo.GetLen()
}

func (n *NormalScheduler) IsOverRefillThreshold() bool {
	return n.queue.refillThreshold <= n.queue.fifo.GetLen()
}
