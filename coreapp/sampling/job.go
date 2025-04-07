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
	jobData        *core.JobData
	jobContext     *core.JobContext
	mitigationInfo *mitig.MitigationInfo
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
	j.setMitigationInfo()
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
	j.mitigationInfo.Mitigated = true
	if j.mitigationInfo.Readout == "pseudo_inverse" {
		zap.L().Debug(fmt.Sprintf("start to do pseudo inverse mitigation"))
		mitig.PseudoInverseMitigation(j.JobData())
	} else {
		zap.L().Debug(fmt.Sprintf("skip pseudo inverse mitigation"))
	}
	return
}

func (j *SamplingJob) IsFinished() bool {
	zap.L().Debug(fmt.Sprintf("checking if job(%s) is finished", j.JobData().ID))
	if j.mitigationInfo.NeedToBeMitigated {
		zap.L().Debug(fmt.Sprintf("job(%s) need to be mitigated", j.JobData().ID))
		return j.mitigationInfo.Mitigated
	} else {
		zap.L().Debug(fmt.Sprintf("job(%s) does not need to be mitigated", j.JobData().ID))
		return j.JobData().Status == core.SUCCEEDED || j.JobData().Status == core.FAILED
	}
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

func (j *SamplingJob) setMitigationInfo() {
	m := mitig.MitigationInfo{}
	if err := json.Unmarshal([]byte(j.JobData().MitigationInfo), &m); err != nil {
		zap.L().Error(fmt.Sprintf("failed to unmarshal MitigationInfo from :%s/reason:%s",
			j.JobData().MitigationInfo, err))
		m.NeedToBeMitigated = false
	} else {
		if m.Readout == "pseudo_inverse" { // TODO: check this condition
			zap.L().Debug(fmt.Sprintf("JobID:%s Need to be mitigated", j.JobData().ID))
			m.NeedToBeMitigated = true
		} else {
			zap.L().Debug(fmt.Sprintf("JobID:%s does not need to be mitigated", j.JobData().ID))
			m.NeedToBeMitigated = false
		}
	}
	m.Mitigated = false
	zap.L().Debug(fmt.Sprintf("set MitigationInfo:%s", j.JobData().MitigationInfo))
	j.mitigationInfo = &m
}
