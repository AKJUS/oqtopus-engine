//go:build unit
// +build unit

package qpu

import (
	"testing"

	"github.com/oqtopus-team/oqtopus-engine/coreapp/common"
	"github.com/stretchr/testify/assert"
)

func TestOneQubitIR(t *testing.T) {
	testQASM, commonErr := common.GetAsset("oneq.qasm")
	assert.Nil(t, commonErr)
	circ, circuitErr := ParseQASM(testQASM)
	assert.Nil(t, circuitErr)

	circIR, irErr := NewCircuitIR(circ.ProgramContext())
	assert.Nil(t, irErr)
	assert.Equal(t, circIR.ProgramIR.Version, "3")
}
func TestRzIR(t *testing.T) {
	testQASM, commonErr := common.GetAsset("rz.qasm")
	assert.Nil(t, commonErr)
	circ, circuitErr := ParseQASM(testQASM)
	assert.Nil(t, circuitErr)

	circIR, irErr := NewCircuitIR(circ.ProgramContext())
	assert.Nil(t, irErr)
	assert.Equal(t, circIR.ProgramIR.Version, "3")
}

func TestXGateIR(t *testing.T) {
	testQASM, commonErr := common.GetAsset("x_gate.qasm")
	assert.Nil(t, commonErr)
	circ, circuitErr := ParseQASM(testQASM)
	assert.Nil(t, circuitErr)

	circIR, irErr := NewCircuitIR(circ.ProgramContext())
	assert.Nil(t, irErr)

	assert.Equal(t, circIR.ProgramIR.Version, "3")

	assert.Equal(t, len(circIR.ProgramIR.QubitAbsNum), 1)
	assert.Equal(t, circIR.ProgramIR.QubitAbsNum[QCbitIdentifier{Name: "q", Index: 0}], 0)
	assert.Equal(t, circIR.ProgramIR.QubitCount, 1)
	assert.Equal(t, circIR.ProgramIR.BitAbsNum[QCbitIdentifier{Name: "c", Index: 0}], 0)
	assert.Equal(t, circIR.ProgramIR.BitCount, 1)

	assert.Equal(
		t,
		circIR.ProgramIR.Statements[2],
		&GateCallStatementIR{
			GateName: "x",
			Operands: []QCbitIdentifier{
				QCbitIdentifier{Name: "q", Index: 0},
			}})

	assert.Equal(
		t,
		circIR.ProgramIR.Statements[3],
		&AssignmentStatementIR{
			Left: QCbitIdentifier{Name: "c", Index: 0},
			Right: MeasureExpressionIR{
				QCbitIdentifier: QCbitIdentifier{Name: "q", Index: 0}}})
}

func TestBellPairIR(t *testing.T) {
	testQASM, commonErr := common.GetAsset("bell_pair.qasm")
	assert.Nil(t, commonErr)
	circ, circuitErr := ParseQASM(testQASM)
	assert.Nil(t, circuitErr)

	circIR, irErr := NewCircuitIR(circ.ProgramContext())
	assert.Nil(t, irErr)

	assert.Equal(t, circIR.ProgramIR.Version, "3")

	assert.Equal(t, len(circIR.ProgramIR.QubitAbsNum), 2)
	assert.Equal(t, circIR.ProgramIR.QubitAbsNum[QCbitIdentifier{Name: "q", Index: 0}], 0)
	assert.Equal(t, circIR.ProgramIR.QubitAbsNum[QCbitIdentifier{Name: "q", Index: 1}], 1)
	assert.Equal(t, circIR.ProgramIR.QubitCount, 2)

	assert.Equal(t, len(circIR.ProgramIR.BitAbsNum), 2)
	assert.Equal(t, circIR.ProgramIR.BitAbsNum[QCbitIdentifier{Name: "c", Index: 0}], 0)
	assert.Equal(t, circIR.ProgramIR.BitAbsNum[QCbitIdentifier{Name: "c", Index: 1}], 1)
	assert.Equal(t, circIR.ProgramIR.BitCount, 2)

	assert.Equal(t, len(circIR.ProgramIR.Statements), 6)
	assert.Equal(t, circIR.ProgramIR.Statements[0], &QuantumDeclarationStatementIR{Identifier: "q", Designator: 2})
	assert.Equal(t, circIR.ProgramIR.Statements[1], &ClassicalDeclarationStatementIR{Identifier: "c", Designator: 2})
	assert.Equal(
		t,
		circIR.ProgramIR.Statements[2],
		&GateCallStatementIR{
			GateName: "h",
			Operands: []QCbitIdentifier{
				QCbitIdentifier{Name: "q", Index: 0},
			}})
	assert.Equal(
		t,
		circIR.ProgramIR.Statements[3],
		&GateCallStatementIR{
			GateName: "cx",
			Operands: []QCbitIdentifier{
				QCbitIdentifier{Name: "q", Index: 0},
				QCbitIdentifier{Name: "q", Index: 1},
			}})

	assert.Equal(
		t,
		circIR.ProgramIR.Statements[4],
		&AssignmentStatementIR{
			Left: QCbitIdentifier{Name: "c", Index: 0},
			Right: MeasureExpressionIR{
				QCbitIdentifier: QCbitIdentifier{Name: "q", Index: 0}}})
	assert.Equal(
		t,
		circIR.ProgramIR.Statements[5],
		&AssignmentStatementIR{
			Left: QCbitIdentifier{Name: "c", Index: 1},
			Right: MeasureExpressionIR{
				QCbitIdentifier: QCbitIdentifier{Name: "q", Index: 1}}})
}
