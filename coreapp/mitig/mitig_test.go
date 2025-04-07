//go:build skip
// +build skip

// The tests in this file are skipped because they are supposed to be run with
// mitigator conatainers.
// mitigator mocks should be used to test these...

package mitig

import (
	"testing"

	"github.com/oqtopus-team/oqtopus-engine/coreapp/core"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestGetLowerBits(t *testing.T) {
	tests := []struct {
		name  string
		input string
		n     int
		want  string
	}{
		{
			name:  "test1",
			input: "101010",
			n:     3,
			want:  "010",
		},
		{
			name:  "test2",
			input: "101010",
			n:     2,
			want:  "10",
		},
		{
			name:  "test3",
			input: "101010",
			n:     1,
			want:  "0",
		},
		{
			name:  "test4",
			input: "101010",
			n:     0,
			want:  "",
		},
		{
			name:  "test5",
			input: "101010",
			n:     6,
			want:  "101010",
		},
		{
			name:  "test6",
			input: "101010",
			n:     7,
			want:  "101010",
		},
		{
			name:  "test7",
			input: "101010",
			n:     8,
			want:  "101010",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getLowerBits(tt.input, tt.n)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewMitigationInfoFromJobData(t *testing.T) {
	tests := []struct {
		name           string
		mitigationInfo string
		want           MitigationInfo
	}{
		{
			name:           "pseudo_inverse mitigation",
			mitigationInfo: `{"readout": "pseudo_inverse"}`,
			want:           MitigationInfo{NeedToBeMitigated: true, Mitigated: false, Readout: "pseudo_inverse"},
		},
		{
			name:           "other readout mitigation",
			mitigationInfo: `{"readout": "other"}`,
			want:           MitigationInfo{NeedToBeMitigated: false, Mitigated: false, Readout: "other"},
		},
		{
			name:           "no readout field",
			mitigationInfo: `{"some_other_field": "value"}`,
			want:           MitigationInfo{NeedToBeMitigated: false, Mitigated: false, Readout: ""},
		},
		{
			name:           "invalid json",
			mitigationInfo: `{"readout": "pseudo_inverse"`, // Missing closing brace
			want:           MitigationInfo{NeedToBeMitigated: false, Mitigated: false, Readout: ""},
		},
		{
			name:           "empty string",
			mitigationInfo: ``,
			want:           MitigationInfo{NeedToBeMitigated: false, Mitigated: false, Readout: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jd := &core.JobData{
				MitigationInfo: tt.mitigationInfo,
				ID:             "test-job-" + tt.name, // Add unique ID for logging clarity
			}
			got := NewMitigationInfoFromJobData(jd)

			// We only check NeedToBeMitigated and Mitigated as Readout might not be set correctly on error
			assert.Equal(t, tt.want.NeedToBeMitigated, got.NeedToBeMitigated, "NeedToBeMitigated mismatch")
			assert.Equal(t, tt.want.Mitigated, got.Mitigated, "Mitigated mismatch")
			// Optionally check Readout only if parsing was expected to succeed
			if tt.name != "invalid json" && tt.name != "empty string" {
				assert.Equal(t, tt.want.Readout, got.Readout, "Readout mismatch")
			}
		})
	}
}

func TestPseudoInverseMitigation(t *testing.T) {
	s := core.SCWithUnimplementedContainer()
	defer s.TearDown()

	jd := core.NewJobData()
	jd.ID = "test_mitigation_job"
	jd.QASM = "OPENQASM 3.0;\ninclude \"stdgates.inc\";\nqubit[2] q;\nh q[0];\ncx q[0], q[1];\n"
	// Result.Counts is expected to filled Process() phase
	jd.Result.Counts = core.Counts{"00": 250, "01": 250, "10": 250, "11": 250}

	// PhsicalVirtualMapping is expected to filled transpile phase
	jd.Result.TranspilerInfo.PhysicalVirtualMapping = map[uint32]uint32{0: 0, 1: 1}

	PseudoInverseMitigation(jd)
	actual := jd.Result.Counts
	expect := core.Counts{"00": 191, "01": 268, "10": 225, "11": 315}

	assert.Equal(t, expect, actual)
}
