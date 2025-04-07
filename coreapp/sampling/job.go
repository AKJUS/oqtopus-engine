package sampling

import (
	"encoding/json"
	"fmt"

	"github.com/oqtopus-team/oqtopus-engine/coreapp/core"
	"github.com/oqtopus-team/oqtopus-engine/coreapp/mitig"
	"go.uber.org/zap"
)

const SAMPLING_JOB = "sampling"

type SamplingJob struct {
	jobData    *core.JobData
	jobContext *core.JobContext
}

func (j *SamplingJob) New(jd *core.JobData, jc *core.JobContext) core.Job {
	return &SamplingJob{
		jobData:    jd,
		jobContext: jc,
	}
}

func (j *SamplingJob) PreProcess() {
	if err := j.preProcessImpl(); err != nil {
		zap.L().Error(fmt.Sprintf("failed to pre-process a job(%s). Reason:%s",
			j.JobData().ID, err.Error()))
		core.SetFailureWithError(j, err)
		return
	}
	return
}

func (j *SamplingJob) preProcessImpl() (err error) {
	err = nil
	jd := j.JobData()
	container := core.GetSystemComponents().Container
	// TODO refactor this part
	// make jobID pool in syscomponent
	err = container.Invoke(
		func(d core.DBManager) error {
			if d.ExistInInnerJobIDSet(jd.ID) {
				return core.ErrorJobIDConflict
			}
			return nil
		})
	if err != nil {
		zap.L().Error(fmt.Sprintf("failed to check the existence of a job(%s). Reason:%s",
			jd.ID, err.Error()))
		return
	}

	if jd.NeedTranspiling() {
		err = container.Invoke(
			func(t core.Transpiler) error {
				return t.Transpile(j)
			})
		if err != nil {
			zap.L().Error(fmt.Sprintf("failed to transpile a job(%s). Reason:%s", jd.ID, err.Error()))
			return
		}
	} else {
		zap.L().Debug(fmt.Sprintf("skip transpiling a job(%s)/Transpiler:%v",
			jd.ID, jd.Transpiler))
	}
	_ = container.Invoke(
		func(d core.DBManager) error {
			d.AddToInnerJobIDSet(jd.ID)
			return nil
		})
	return
}

func (j *SamplingJob) Process() {
	c := core.GetSystemComponents().Container
	err := c.Invoke(
		func(q core.QPUManager) error {
			return q.Send(j)
		})
	if err != nil {
		zap.L().Error(fmt.Sprintf("failed to send a job(%s) to QPU. Reason:%s", j.JobData().ID, err.Error()))
		j.JobData().Status = core.FAILED
	}
	zap.L().Debug(fmt.Sprintf("finished to process a job(%s)/status:%s", j.JobData().ID, j.JobData().Status))
}

func (j *SamplingJob) PostProcess() {
	m := mitig.MitigationInfo{}
	err := json.Unmarshal([]byte(j.JobData().MitigationInfo), &m)
	if err != nil {
		zap.L().Error(fmt.Sprintf("failed to unmarshal MitigationInfo from :%s/reason:%s",
			j.JobData().MitigationInfo, err))
	}
	if m.Readout == "pseudo_inverse" {
		zap.L().Debug(fmt.Sprintf("start to do pseudo inverse mitigation"))
		mitig.PseudoInverseMitigation(j.JobData())
	} else {
		zap.L().Debug(fmt.Sprintf("skip pseudo inverse mitigation"))
	}
	return
}

func (j *SamplingJob) IsFinished() bool {
	return j.JobData().Status == core.SUCCEEDED || j.JobData().Status == core.FAILED
}

func (j *SamplingJob) JobData() *core.JobData {
	return j.jobData
}

func (j *SamplingJob) JobType() string {
	return SAMPLING_JOB
}

func (j *SamplingJob) JobContext() *core.JobContext {
	return j.jobContext
}

func (j *SamplingJob) UpdateJobData(jd *core.JobData) {
	j.jobData = jd
}

func (j *SamplingJob) Clone() core.Job {
	cloned := &SamplingJob{
		jobData:    j.jobData.Clone(),
		jobContext: j.jobContext,
	}
	return cloned
}
