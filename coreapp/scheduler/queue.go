package scheduler

import (
	"fmt"

	conq "github.com/enriquebris/goconcurrentqueue"
	"github.com/oqtopus-team/oqtopus-engine/coreapp/core"
	"go.uber.org/zap"
)

type queueChan chan *jobInScheduler

type fifo interface {
	Enqueue(*jobInScheduler) error
	Dequeue() (*jobInScheduler, error)
	DequeueOrWaitForNextElement() (*jobInScheduler, error)
	Get(index int) (*jobInScheduler, error)
	GetLen() int
	Remove(index int) error
}

type conqFIFO struct {
	conq.FIFO
}

func newConqFIFO() *conqFIFO {
	return &conqFIFO{
		FIFO: *conq.NewFIFO(),
	}
}

func (c *conqFIFO) Enqueue(js *jobInScheduler) error {
	return c.FIFO.Enqueue(js)
}

func (c *conqFIFO) Dequeue() (*jobInScheduler, error) {
	tmp, err := c.FIFO.Dequeue()
	if err != nil {
		return nil, err
	}
	return tmp.(*jobInScheduler), nil
}

func (c *conqFIFO) DequeueOrWaitForNextElement() (*jobInScheduler, error) {
	tmp, err := c.FIFO.DequeueOrWaitForNextElement()
	if err != nil {
		return nil, err
	}
	return tmp.(*jobInScheduler), nil
}

func (c *conqFIFO) Get(index int) (*jobInScheduler, error) {
	tmp, err := c.FIFO.Get(index)
	if err != nil {
		return nil, err
	}
	return tmp.(*jobInScheduler), nil
}

func (c *conqFIFO) GetLen() int {
	return c.FIFO.GetLen()
}

func (c *conqFIFO) Remove(index int) error {
	return c.FIFO.Remove(index)
}

type NormalQueue struct {
	fifo            fifo
	maxSize         int
	refillThreshold int
	queueChan       queueChan
	cancelChan      chan struct{}
}

// TODO: use rungroup
func (n *NormalQueue) Setup(conf *core.Conf) error {
	n.refillThreshold = conf.QueueRefillThreshold
	n.maxSize = conf.QueueMaxSize
	n.fifo = newConqFIFO()
	n.queueChan = make(queueChan)
	n.cancelChan = make(chan struct{})
	go func() {
		defer close(n.cancelChan)
		for {
			var jis *jobInScheduler
			select {
			case <-n.cancelChan:
				return
			case jis = <-n.queueChan:
			}
			jd := jis.job.JobData()
			if n.maxSize <= n.fifo.GetLen() {
				zap.L().Info(fmt.Sprintf("Failed to put %s. Normal Queue is full.", jd.ID))
				continue
			}
			zap.L().Debug(fmt.Sprintf("Putting %s to normalQueue", jd.ID))
			err := n.fifo.Enqueue(jis)
			if err != nil {
				zap.L().Error(
					fmt.Sprintf("Failed to put %s to normalQueue. Reason:%s", jd.ID, err))
			}
		}
	}()
	return nil
}

func (n *NormalQueue) TearDown() {
	n.cancelChan <- struct{}{}
}

// wait until the next elements gets enqueued
func (n *NormalQueue) Dequeue(wait bool) (jis *jobInScheduler, err error) {
	var tmp interface{} // TODO: use generic type?
	jis = nil
	err = nil
	if wait {
		tmp, err = n.fifo.DequeueOrWaitForNextElement()
	} else {
		tmp, err = n.fifo.Dequeue()
	}
	if err != nil {
		zap.L().Debug("no job in NormalQueue.", zap.Error(err))
		return
	}
	jis = tmp.(*jobInScheduler)
	zap.L().Debug(fmt.Sprintf("Dequeued job:%s", jis.job.JobData().ID))
	return
}

func (n *NormalQueue) Delete(jobID string) error {
	zap.L().Debug(fmt.Sprintf("deleting %s to normalQueue", jobID))
	var idx int
	var err error

	idx, err = n.getIdx(jobID)
	if err != nil {
		zap.L().Info(fmt.Sprintf("Failed to Delete %s. Reason:%s", jobID, err))
		return err
	}
	err = n.fifo.Remove(idx)
	if err != nil {
		zap.L().Error(fmt.Sprintf("Failed to remove idx:%d. Reason:%s", idx, err))
		return err
	}
	return nil
}

func (n *NormalQueue) IsOverRefillThreshold() bool {
	return n.refillThreshold <= n.fifo.GetLen()
}

func (n *NormalQueue) GetCurrentSize() int {
	return n.fifo.GetLen()
}

func (n *NormalQueue) getIdx(jobID string) (int, error) {
	for i := 0; i < n.fifo.GetLen(); i++ {
		js, err := n.fifo.Get(i)
		if err == nil {
			if js.job.JobData().ID == jobID {
				return i, nil
			}
		}
	}
	return 0, fmt.Errorf("No entry")
}
