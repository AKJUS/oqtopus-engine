package qpu

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/oqtopus-team/oqtopus-engine/coreapp/core"

	"go.uber.org/zap"
)

const successPropability float32 = 0.9

var source rand.Source = rand.NewSource(time.Now().UnixNano())
var randGenerator *rand.Rand = rand.New(source)

const DummyDeviceName = "DummyQPU"
const DummyProviderName = "DummyProvider"

type DummyQPU struct {
	deviceSetting *DeviceSetting

	EnableDummyQPUTimeInsertion bool
	DummyQPUTime                int
}

func (d *DummyQPU) Setup(conf *core.Conf) error {
	zap.L().Debug("setting up Dummy-QPU")
	d.deviceSetting = NewDeviceSetting()
	d.EnableDummyQPUTimeInsertion = conf.EnableDummyQPUTimeInsertion
	d.DummyQPUTime = conf.DummyQPUTime
	return nil
}

func (d *DummyQPU) Send(inputJob core.Job) error {
	outputJobData := core.CloneJobData(inputJob.JobData())

	zap.L().Info("[Dummy] starting QPU execution")
	if d.EnableDummyQPUTimeInsertion {
		zap.L().Debug(fmt.Sprintf("[Dummy] waiting %d seconds for QPU execution", d.DummyQPUTime))
		<-time.After(time.Duration(d.DummyQPUTime) * time.Second)
	} else {
		zap.L().Debug("[Dummy] no waiting for QPU execution")
	}
	zap.L().Info("[Dummy] finished QPU execution")
	outputJobData.Result.Message = successOrFailure()
	jm := core.GetJobManager()
	job, err := jm.NewJobFromJobData(outputJobData, inputJob.JobContext())
	if err != nil {
		return err
	}
	job.JobContext().DBChan <- job
	return nil
}

func (d *DummyQPU) Validate(qasm string) error {
	return nil //adhoc
	// return circuitValidate(qasm, d.deviceSetting)
}

func (d *DummyQPU) GetDeviceInfo() *core.DeviceInfo {
	return &core.DeviceInfo{
		DeviceName:   DummyDeviceName,
		ProviderName: DummyProviderName,
		Type:         "DummyQPU",
		Status:       core.Available,
		MaxQubits:    10000,
		MaxShots:     10000,
	}
}

func successOrFailure() string {
	if randGenerator.Intn(100) < int(100*successPropability) {
		return "dummy success result"
	}
	return "dummy failure result"
}

type QMTQPU struct {
	agent             QMTAgent
	deviceSetting     *DeviceSetting
	connected         bool
	currentDeviceInfo *core.DeviceInfo

	EnableDummyQPUTimeInsertion bool
	DummyQPUTime                int
}

func (q *QMTQPU) Setup(conf *core.Conf) error {
	zap.L().Debug("Setting up QMT QPU")
	q.EnableDummyQPUTimeInsertion = conf.EnableDummyQPUTimeInsertion
	q.DummyQPUTime = conf.DummyQPUTime

	// TODO remove Device Setting
	ds, err := LoadDeviceSetting(conf.DeviceSettingPath)
	if err != nil {
		zap.L().Error(fmt.Sprintf("Failed to load a device setting. Reason:%s", err))
		return err
	}
	switch ds.DeviceName {
	case "wako", "handai":
		q.agent = NewQMTAgent()
		zap.L().Debug(fmt.Sprintf("Setting up QMT QPU for %s", ds.DeviceName))
	default:
		return fmt.Errorf("unknown device name:%s", ds.DeviceName)
	}
	if err := q.agent.Setup(); err != nil {
		zap.L().Error(fmt.Sprintf("failed to setup QMT QPU/reason:%s", err))
		return err
	}
	q.deviceSetting = ds
	q.connected = false
	if !conf.DisableStartDevicePolling {
		q.startDevicePolling()
	}
	q.currentDeviceInfo = &core.DeviceInfo{
		Status: core.Unavailable,
	}
	return nil
}

func (q *QMTQPU) Validate(qasm string) error {
	return nil //adhoc
	//return circuitValidate(qasm, q.deviceSetting)
}

func (q *QMTQPU) Send(j core.Job) error {
	var err error
	jd := j.JobData()
	zap.L().Info("Starting QMT QPU execution of Job ID:" + jd.ID)

	if !q.GetConnected() {
		err := fmt.Errorf("QMT QPU is not connected")
		msg := core.SetFailureWithError(j, err)
		zap.L().Info(msg)
		return err
	}
	zap.L().Debug(fmt.Sprintf("Job ID:%s is processing", jd.ID))
	err = q.agent.CallJob(j)

	if err != nil {
		zap.L().Error(fmt.Sprintf("Failed to Call the job (%s) in %s. Reeason:%s",
			jd.ID, q.agent.GetAddress(), err))
		msg := core.SetFailureWithError(j, err)
		zap.L().Info(msg)
		return err
	}
	zap.L().Debug(fmt.Sprintf("Job ID:%s is processed/status: %s", jd.ID, jd.Status))
	jd.Ended = strfmt.DateTime(time.Now())
	return nil
}

func (q *QMTQPU) GetDeviceInfo() *core.DeviceInfo {
	return q.currentDeviceInfo
}

func (q *QMTQPU) GetConnected() bool {
	return q.connected
}

func (q *QMTQPU) startDevicePolling() {
	go func() {
		t := time.NewTicker(time.Duration(q.deviceSetting.PollingPeriod) * time.Second)
		zap.L().Debug("Starting Device Polling")
		q.startCleanUpGoroutine(t)
		for {
			di, err := q.agent.CallDeviceInfo()
			if err != nil {
				zap.L().Error(fmt.Sprintf("Failed to call device info. Reason:%s", err))
				q.currentDeviceInfo = &core.DeviceInfo{Status: core.Unavailable}
				q.connected = false
			} else {
				q.currentDeviceInfo = di
				q.connected = true
			}
			zap.L().Debug(fmt.Sprintf(
				"Waiting %d seconds for the next device polling to %s",
				q.deviceSetting.PollingPeriod, q.agent.GetAddress()))
			<-t.C
		}
	}()
}

// TODO use run Group
func (q *QMTQPU) startCleanUpGoroutine(t *time.Ticker) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		t.Stop()
		q.agent.Close()
	}()
}
