//go:build unit
// +build unit

package qpu

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/oqtopus-team/oqtopus-engine/coreapp/common"
	"github.com/oqtopus-team/oqtopus-engine/coreapp/core"
)

const testQASM = "OPENQASM 3;qubit[1] q;bit[1] c;x q[0];c[0] = measure q[0];"
const testTranspiledQASM = "OPENQASM 3.0;include \"stdgates.inc\";qreg q[64];creg c[64];" +
	"x q[0];c[0] = measure q[0];"

func TestGatewayQPUSend(t *testing.T) { // Renamed from TestQMTQPUSend
	tests := []struct {
		name                string
		connected           bool
		agent               GatewayAgent // Renamed from QMTAgent
		jobID               string
		inputQASM           string
		transpiledInputQASM string
		sentToQPU           bool
		wantMessage         string
		wantErr             *regexp.Regexp
	}{
		{
			name:        "unconnected failure",
			connected:   false,
			agent:       &MockGatewayAgent{}, // Renamed from MockQMTAgent
			jobID:       "test_unconnected_failure",
			sentToQPU:   false,
			inputQASM:   testQASM,
			wantMessage: "Gateway QPU is not connected",                     // Renamed from QMT
			wantErr:     regexp.MustCompile("Gateway QPU is not connected"), // Renamed from QMT
		},
		{
			name:        "call job failure",
			connected:   true,
			agent:       &MockGatewayAgentError{}, // Renamed from MockQMTAgentError
			jobID:       "test_call_job_failure",
			sentToQPU:   true,
			inputQASM:   testQASM,
			wantMessage: "failed to call job",
			wantErr:     regexp.MustCompile("failed to call job"),
		},
	}
	core.ResetSetting()
	core.RegisterSetting("gateway", NewDefaultGatewayAgentSetting()) // Renamed from "qmt", NewDefaultQMTAgentSetting

	s := core.SCWithUnimplementedContainer()
	defer s.TearDown()
	jm, err := core.NewJobManager(&core.NormalJob{})
	assert.Nil(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsPath, programErr := common.GetAssetAbsPath("unit_test_device_setting.toml")
			if programErr != nil {
				t.Fatal(programErr)
			}
			conf := &core.Conf{
				DeviceSettingPath:         dsPath,
				DisableStartDevicePolling: true,
			}
			gatewayQPU := &GatewayQPU{}        // Renamed from QMTQPU
			setupErr := gatewayQPU.Setup(conf) // Renamed from qmtQPU
			assert.Nil(t, setupErr)
			gatewayQPU.agent = tt.agent         // Renamed from qmtQPU
			gatewayQPU.connected = tt.connected // Renamed from qmtQPU

			jd := core.NewJobData()
			jd.ID = tt.jobID
			jd.QASM = tt.inputQASM
			jd.TranspiledQASM = tt.transpiledInputQASM
			jd.Transpiler = core.DEFAULT_TRANSPILER_CONFIG()
			jd.JobType = core.NORMAL_JOB
			jc, err := core.NewJobContext()
			assert.Nil(t, err)
			nj, err := jm.NewJobFromJobData(jd, jc)
			assert.Nil(t, err)

			sendErr := gatewayQPU.Send(nj) // Renamed from qmtQPU
			if sendErr != nil {
				assert.Regexp(t, tt.wantErr, sendErr)
			}

			assert.True(t, time.Time(jd.Ended).After(time.Time(jd.Created)))
			assert.Equal(t, tt.wantMessage, jd.Result.Message)
		})
	}
}

type MockGatewayAgent struct{} // Renamed from MockQMTAgent

func (m *MockGatewayAgent) Setup() error { // Renamed from MockQMTAgent
	return nil
}

func (m *MockGatewayAgent) CallJob(j core.Job) error { // Renamed from MockQMTAgent
	return nil
}

func (m *MockGatewayAgent) CallDeviceInfo() (*core.DeviceInfo, error) { // Renamed from MockQMTAgent
	return &core.DeviceInfo{
		DeviceName: "mock_gateway_client", // Renamed from "mock_qmt_client"
	}, nil
}

func (m *MockGatewayAgent) Reset() {} // Renamed from MockQMTAgent

func (m *MockGatewayAgent) Close() {} // Renamed from MockQMTAgent

func (m *MockGatewayAgent) GetAddress() string { // Renamed from MockQMTAgent
	return "dummy_address"
}

type MockGatewayAgentError struct { // Renamed from MockQMTAgentError
	MockGatewayAgent // Renamed from MockQMTAgent
}

func (m *MockGatewayAgentError) CallJob(j core.Job) error { // Renamed from MockQMTAgentError
	jd := j.JobData()
	jd.Result.Message = "failed to call job"
	return nil
}
