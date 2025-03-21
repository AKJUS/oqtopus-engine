//go:build unit
// +build unit

package scheduler

import (
	"github.com/oqtopus-team/oqtopus-engine/coreapp/core"
	"github.com/stretchr/testify/assert"

	"testing"
)

type TestFIFO struct {
	conqFIFO
	queuedChan chan struct{}
}

func newTestFIFO(queuedChan chan struct{}) *TestFIFO {
	return &TestFIFO{
		conqFIFO:   *newConqFIFO(),
		queuedChan: queuedChan,
	}
}

func (t *TestFIFO) Enqueue(js *jobInScheduler) error {
	err := t.FIFO.Enqueue(js)
	t.queuedChan <- struct{}{}
	return err
}

func setUpTestNormalQueue(queuedChan chan struct{}) *NormalQueue {
	n := &NormalQueue{}
	conf := &core.Conf{QueueMaxSize: 1000}
	n.Setup(conf)
	n.fifo = newTestFIFO(queuedChan)
	return n
}

func tearDownTestNormalQueue(n *NormalQueue) {
	close(n.fifo.(*TestFIFO).queuedChan)
	n.TearDown()
}

func TestPutNormalQueue(t *testing.T) {
	s := core.SCWithUnimplementedContainer()
	defer s.TearDown()
	queuedChan := make(chan struct{})
	n := setUpTestNormalQueue(queuedChan)
	defer tearDownTestNormalQueue(n)

	n.queueChan <- newjobInScheduler(t, "test1")
	<-queuedChan
	assert.Equal(t, 1, n.fifo.GetLen())
	js, err := n.Dequeue(false)
	assert.Nil(t, err)
	assert.Equal(t, js.job.JobData().ID, "test1")
}

func TestNormalQueueDelete(t *testing.T) {
	s := core.SCWithUnimplementedContainer()
	defer s.TearDown()
	queuedChan := make(chan struct{})
	n := setUpTestNormalQueue(queuedChan)
	defer tearDownTestNormalQueue(n)

	n.queueChan <- newjobInScheduler(t, "test1")
	<-queuedChan
	assert.Equal(t, n.fifo.GetLen(), 1)
	n.queueChan <- newjobInScheduler(t, "test2")
	<-queuedChan
	assert.Equal(t, n.fifo.GetLen(), 2)
	n.queueChan <- newjobInScheduler(t, "test3")
	<-queuedChan
	assert.Equal(t, n.fifo.GetLen(), 3)
	n.queueChan <- newjobInScheduler(t, "test4")
	<-queuedChan
	assert.Equal(t, n.fifo.GetLen(), 4)

	n.Delete("test3")

	assert.Equal(t, n.fifo.GetLen(), 3)

	var jwg *jobInScheduler
	var err error

	jwg, err = n.Dequeue(false)
	assert.Nil(t, err)
	assert.Equal(t, jwg.job.JobData().ID, "test1")

	jwg, err = n.Dequeue(false)
	assert.Nil(t, err)
	assert.Equal(t, jwg.job.JobData().ID, "test2")

	jwg, err = n.Dequeue(false)
	assert.Nil(t, err)
	assert.Equal(t, jwg.job.JobData().ID, "test4")

	jwg, err = n.Dequeue(false)
	assert.EqualError(t, err, "empty queue")
	assert.Nil(t, jwg)
}

func newjobInScheduler(t *testing.T, id string) *jobInScheduler {
	jm, err := core.NewJobManager(&core.NormalJob{})
	assert.Nil(t, err)
	jc, err := core.NewJobContext()
	assert.Nil(t, err)
	nj, err := jm.NewJobFromJobData(&core.JobData{ID: id}, jc)
	assert.Nil(t, err)
	return &jobInScheduler{
		job: nj,
	}
}
