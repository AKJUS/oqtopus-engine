package qpu

import (
	"fmt"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/oqtopus-team/oqtopus-engine/coreapp/core/parser"
	"go.uber.org/zap"
)

type QASMStatementType struct {
	Name string
}

func (q *QASMStatementType) String() string {
	return q.Name
}

type QASMGateType struct {
	Name string
}

func (q *QASMGateType) String() string {
	return q.Name
}

type CircuitIR struct {
	ProgramIR *ProgramIR

	parser.Baseqasm3ParserListener

	tmpStatementIR         StatementIR
	tmpDegsinator          int
	tmpQCbitIdentifiers    []QCbitIdentifier
	tmpMeasureExpressionIR MeasureExpressionIR
	tmpExpressionList      string
}

func NewCircuitIR(pc *parser.ProgramContext) (circIR *CircuitIR, err error) {
	circIR = &CircuitIR{}
	err = nil
	defer func() {
		if recErr := recover(); recErr != nil {
			msg := fmt.Sprintf("failed to generate IR.")
			zap.L().Error(msg)
			zap.L().Debug(fmt.Sprintf("reason:%s\nqasm:\n%s", recErr, pc.GetText()))
			circIR = &CircuitIR{}
			err = fmt.Errorf(msg)
		}
	}()
	antlr.ParseTreeWalkerDefault.Walk(circIR, pc)
	return
}

func (c *CircuitIR) EnterProgram(ctx *parser.ProgramContext) {
	c.ProgramIR = &ProgramIR{
		QubitCount:  0,
		QubitAbsNum: make(map[QCbitIdentifier]int),
		BitCount:    0,
		BitAbsNum:   make(map[QCbitIdentifier]int),
	}
}

func (c *CircuitIR) EnterVersion(ctx *parser.VersionContext) {
	c.ProgramIR.Version = ctx.VersionSpecifier().GetText()
}

func (c *CircuitIR) EnterQuantumDeclarationStatement(ctx *parser.QuantumDeclarationStatementContext) {
	c.tmpStatementIR = &QuantumDeclarationStatementIR{Identifier: ctx.Identifier().GetText()}
}

func (c *CircuitIR) ExitQuantumDeclarationStatement(ctx *parser.QuantumDeclarationStatementContext) {
	newQDStatement := c.tmpStatementIR.(*QuantumDeclarationStatementIR)
	newQDStatement.Designator = c.tmpDegsinator
	for i := 0; i < newQDStatement.Designator; i++ {
		qi := QCbitIdentifier{
			Name:  newQDStatement.Identifier,
			Index: i,
		}
		c.ProgramIR.QubitAbsNum[qi] = c.ProgramIR.QubitCount
		c.ProgramIR.QubitCount++
	}
	c.ProgramIR.Statements = append(c.ProgramIR.Statements, newQDStatement)
}

func (c *CircuitIR) EnterDesignator(ctx *parser.DesignatorContext) {
	des, err := strconv.Atoi(ctx.Expression().GetText())
	if err != nil {
		panic(err)
	}
	c.tmpDegsinator = des
}

func (c *CircuitIR) EnterClassicalDeclarationStatement(ctx *parser.ClassicalDeclarationStatementContext) {
	c.tmpStatementIR = &ClassicalDeclarationStatementIR{Identifier: ctx.Identifier().GetText()}
}

func (c *CircuitIR) ExitClassicalDeclarationStatement(ctx *parser.ClassicalDeclarationStatementContext) {
	newCDStatement := c.tmpStatementIR.(*ClassicalDeclarationStatementIR)
	newCDStatement.Designator = c.tmpDegsinator
	for i := 0; i < newCDStatement.Designator; i++ {
		bi := QCbitIdentifier{
			Name:  newCDStatement.Identifier,
			Index: i,
		}
		c.ProgramIR.BitAbsNum[bi] = c.ProgramIR.BitCount
		c.ProgramIR.BitCount++
	}
	c.ProgramIR.Statements = append(c.ProgramIR.Statements, newCDStatement)
}

func (c *CircuitIR) EnterGateCallStatement(ctx *parser.GateCallStatementContext) {
	c.tmpStatementIR = &GateCallStatementIR{GateName: ctx.Identifier().GetText()}
}

func (c *CircuitIR) ExitGateCallStatement(ctx *parser.GateCallStatementContext) {
	c.tmpStatementIR.(*GateCallStatementIR).Operands = c.tmpQCbitIdentifiers
	c.tmpStatementIR.(*GateCallStatementIR).ExpList = c.tmpExpressionList
	c.ProgramIR.Statements = append(c.ProgramIR.Statements, c.tmpStatementIR)
}

func (c *CircuitIR) EnterGateOperandList(ctx *parser.GateOperandListContext) {
	c.tmpQCbitIdentifiers = []QCbitIdentifier{}
}

func (c *CircuitIR) EnterExpressionList(ctx *parser.ExpressionListContext) {
	c.tmpExpressionList = ""
}

func (c *CircuitIR) ExitExpressionList(ctx *parser.ExpressionListContext) {
	c.tmpExpressionList = ctx.GetText()
}

func (c *CircuitIR) EnterIndexedIdentifier(ctx *parser.IndexedIdentifierContext) {
	ind, err := strconv.Atoi(ctx.IndexOperator(0).(*parser.IndexOperatorContext).Expression(0).GetText())
	if err != nil {
		panic(err)
	}
	c.tmpQCbitIdentifiers = append(
		c.tmpQCbitIdentifiers,
		QCbitIdentifier{
			Name:  ctx.Identifier().GetText(),
			Index: ind})
}

func (c *CircuitIR) EnterAssignmentStatement(ctx *parser.AssignmentStatementContext) {
	c.tmpStatementIR = &AssignmentStatementIR{}
	c.tmpQCbitIdentifiers = []QCbitIdentifier{}
}

func (c *CircuitIR) EnterMeasureExpression(ctx *parser.MeasureExpressionContext) {

}

func (c *CircuitIR) ExitMeasureExpression(ctx *parser.MeasureExpressionContext) {
	c.tmpMeasureExpressionIR = MeasureExpressionIR{
		QCbitIdentifier: c.tmpQCbitIdentifiers[1],
	}
}

func (c *CircuitIR) ExitAssignmentStatement(ctx *parser.AssignmentStatementContext) {
	c.tmpStatementIR.(*AssignmentStatementIR).Left = c.tmpQCbitIdentifiers[0]
	c.tmpStatementIR.(*AssignmentStatementIR).Right = c.tmpMeasureExpressionIR
	c.ProgramIR.Statements = append(c.ProgramIR.Statements, c.tmpStatementIR)
}

type ProgramIR struct {
	Version    string
	Statements []StatementIR

	QubitCount  int
	QubitAbsNum map[QCbitIdentifier]int // Get absolute qubit number
	BitCount    int
	BitAbsNum   map[QCbitIdentifier]int // Get absolute bit number

}

type QCbitIdentifier struct {
	Name  string
	Index int
}

type MeasureExpressionIR struct {
	QCbitIdentifier QCbitIdentifier
}

type StatementIR interface {
	String() string
	IsStatementIR()
}

type QuantumDeclarationStatementIR struct {
	Identifier string
	Designator int
}

func (QuantumDeclarationStatementIR) IsStatementIR() {}
func (QuantumDeclarationStatementIR) String() string {
	return "QuantumDecalrationStatementIR"
}

type ClassicalDeclarationStatementIR struct {
	Identifier string
	Designator int
}

func (ClassicalDeclarationStatementIR) IsStatementIR() {}
func (ClassicalDeclarationStatementIR) String() string {
	return "ClassicalDecalrationStatementIR"
}

type GateCallStatementIR struct {
	GateName string
	Operands []QCbitIdentifier
	ExpList  string
}

func (GateCallStatementIR) IsStatementIR() {}
func (GateCallStatementIR) String() string {
	return "GateCallStatementIR"
}

type AssignmentStatementIR struct {
	Left  QCbitIdentifier
	Right MeasureExpressionIR
}

func (AssignmentStatementIR) IsStatementIR() {}
func (AssignmentStatementIR) String() string {
	return "AssignmentStatementIR"
}

type Designator struct {
	Expression int
}
