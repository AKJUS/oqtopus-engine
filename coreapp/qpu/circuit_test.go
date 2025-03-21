//go:build unit
// +build unit

package qpu

import (
	"strconv"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/oqtopus-team/oqtopus-engine/coreapp/common"
	"github.com/oqtopus-team/oqtopus-engine/coreapp/core"
	"github.com/stretchr/testify/assert"
)

var testDeviceSetting *DeviceSetting = &DeviceSetting{
	QASMSupport: NewQasmSupport(),
}

func TestCircuitValidate(t *testing.T) {
	s := core.SCWithUnimplementedContainer()
	defer s.TearDown()
	maxQubits := s.GetDeviceInfo().MaxQubits
	assert.Equal(t, maxQubits, core.MockMaxQubits)

	tests := []struct {
		name          string
		qasm          string
		deviceSetting *DeviceSetting
		wantErrorMsg  string
	}{
		{
			name:          "not qasm statement",
			qasm:          "hoge",
			deviceSetting: testDeviceSetting,
			wantErrorMsg:  "line 1:4 no viable alternative at input 'hoge'",
		},
		{
			name:          "bad qubit declaration",
			qasm:          "qubit[3]",
			deviceSetting: testDeviceSetting,
			wantErrorMsg:  "canceled to parse",
		},
		{
			name:          "qubit declaration",
			qasm:          "qubit[3] a;",
			deviceSetting: testDeviceSetting,
			wantErrorMsg:  "",
		},
		{
			name:          "full size qubits",
			qasm:          "qubit[" + strconv.Itoa(maxQubits) + "] a;",
			deviceSetting: testDeviceSetting,
			wantErrorMsg:  "",
		},
		{
			name:          "too many qubits",
			qasm:          "qubit[" + strconv.Itoa(maxQubits+1) + "] a;",
			deviceSetting: testDeviceSetting,
			wantErrorMsg: "Too many quibits in your circuit. We only have " +
				strconv.Itoa(maxQubits) + " qubits.",
		},
		{
			name:          "gate call",
			qasm:          "h a[0];",
			deviceSetting: testDeviceSetting,
			wantErrorMsg:  "",
		},
		{
			name: "allow and deny list",
			qasm: "if(c0==1) z q[2];",
			deviceSetting: &DeviceSetting{
				QASMSupport: &QASMSupport{
					AllowList: &QASMFilter{
						Enabled: true,
						Statements: []*QASMStatementType{
							&QASMStatementType{Name: "gate_call"},
						},
					},
					DenyList: &QASMFilter{
						Enabled: true,
						Statements: []*QASMStatementType{
							&QASMStatementType{Name: "if"},
						},
					},
				},
			},
			wantErrorMsg: "statement:ifStatement is not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := circuitValidate(tt.qasm, tt.deviceSetting)
			if tt.wantErrorMsg == "" {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, err.Error(), tt.wantErrorMsg)
			}
		})
	}
}

func TestParseBellPair(t *testing.T) {
	testQASM, commonErr := common.GetAsset("bell_pair.qasm")
	assert.Nil(t, commonErr)
	circ, circErr := ParseQASM(testQASM)
	assert.Nil(t, circErr)

	target := heredoc.Doc(`
	(program (version OPENQASM 3 ;) 
	(statement (quantumDeclarationStatement 
	(qubitType qubit (designator [ (expression 2) ])) q ;)) 
	
	(statement (classicalDeclarationStatement 
	(scalarType bit (designator [ (expression 2) ])) c ;)) 

	(statement (gateCallStatement 
	h (gateOperandList (gateOperand (indexedIdentifier q (indexOperator [ (expression 0) ])))) ;)) 

	(statement (gateCallStatement 
	cx (gateOperandList 
	(gateOperand (indexedIdentifier q (indexOperator [ (expression 0) ]))) , 
	(gateOperand (indexedIdentifier q (indexOperator [ (expression 1) ])))) ;)) 
	
	(statement (assignmentStatement 
	(indexedIdentifier c (indexOperator [ (expression 0) ])) = 
	(measureExpression measure (gateOperand (indexedIdentifier q (indexOperator [ (expression 0) ])))) ;)) 
	
	(statement (assignmentStatement 
	(indexedIdentifier c (indexOperator [ (expression 1) ])) = 
	(measureExpression measure (gateOperand (indexedIdentifier q (indexOperator [ (expression 1) ])))) ;)) 
	<EOF>)`)
	assert.Equal(t, circ.stringTree, strings.Replace(target, "\n", "", -1))
}

func TestCheckResource(t *testing.T) {
	tests := []struct {
		name         string
		qasm         string
		wantErrorMsg string
	}{
		{
			name: "valid qasm",
			qasm: heredoc.Doc(`
				OPENQASM 3;
				qubit[2] q;
				h q[0];
				cx q[0], q[1];
				measure q[0] -> c[0];
				measure q[1] -> c[1];
			`),
			wantErrorMsg: "",
		},
		{
			name: "too many qubits",
			qasm: heredoc.Doc(`
				OPENQASM 3;
				qubit[3] q;
				h q[0];
				cx q[0], q[1];
				measure q[0] -> c[0];
				measure q[1] -> c[1];
			`),
			wantErrorMsg: "Too many quibits in your circuit. We only have 2 qubits.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			circ, circErr := ParseQASM(tt.qasm)
			assert.Nil(t, circErr)
			err := checkResource(circ, 2)
			if tt.wantErrorMsg == "" {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, err.Error(), tt.wantErrorMsg)
			}
		})
	}
}
