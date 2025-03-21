package qpu

import (
	"fmt"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/oqtopus-team/oqtopus-engine/coreapp/common"
	"github.com/oqtopus-team/oqtopus-engine/coreapp/core"
	"github.com/oqtopus-team/oqtopus-engine/coreapp/core/parser"
	"go.uber.org/zap"
)

type Circuit struct {
	programContext *parser.ProgramContext
	ruleNames      []string
	stringTree     string
}

func (c *Circuit) ProgramContext() *parser.ProgramContext {
	return c.programContext
}

func circuitValidate(qasm string, ds *DeviceSetting) error {
	if qasm == "" {
		msg := "no input qasm"
		zap.L().Info(msg)
		return fmt.Errorf(msg)
	}
	circ, err := ParseQASM(qasm)
	if err != nil {
		zap.L().Info(err.Error())
		return err
	}
	if err := validateStatements(circ, ds.QASMSupport); err != nil {
		zap.L().Info(err.Error())
		return err
	}
	di := core.GetSystemComponents().GetDeviceInfo()
	if di.Status != core.Available {
		msg := fmt.Sprintf("device is not available. status:%s", di.Status)
		zap.L().Info(msg)
		return fmt.Errorf(msg)
	} else {
		if err := checkResource(circ, di.MaxQubits); err != nil {
			zap.L().Info(err.Error())
			return err
		}
	}
	return nil
}

func ParseQASM(qasm string) (circ *Circuit, err error) {
	if qasm == "" {
		msg := "no input qasm"
		return nil, fmt.Errorf(msg)
	}
	circ = &Circuit{}
	err = nil

	input := antlr.NewInputStream(qasm)
	lexer := parser.Newqasm3Lexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.Newqasm3Parser(stream)
	p.SetErrorHandler(antlr.NewBailErrorStrategy())
	p.RemoveErrorListeners()
	el := &qasmErrorListener{}
	p.AddErrorListener(el)
	defer func() {
		if recErr := recover(); recErr != nil {
			var userErrorMsg, devErrorMsg string
			if el.msgBuffer != "" {
				userErrorMsg = el.msgBuffer
				devErrorMsg = "[qasmErrorListener]" + userErrorMsg
			} else {
				switch recErr.(type) {
				case *antlr.ParseCancellationException:
					userErrorMsg = "canceled to parse"
					devErrorMsg = "[*antlr.ParseCancellationException]" + userErrorMsg
				default:
					userErrorMsg = "failed to parse"
					devErrorMsg = "[Unknown Error]" + userErrorMsg
				}
			}
			zap.L().Info(devErrorMsg)
			zap.L().Debug(fmt.Sprintf("qasm:\n%s", qasm))
			err = fmt.Errorf(userErrorMsg)
		}
	}()
	parsed := p.Program()
	circ.programContext = parsed.(*parser.ProgramContext)
	circ.ruleNames = p.RuleNames
	circ.stringTree = circ.programContext.ToStringTree(circ.ruleNames, p)
	return
}

type qasmErrorListener struct {
	antlr.DefaultErrorListener
	msgBuffer string
}

func (q *qasmErrorListener) SyntaxError(_ antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	q.msgBuffer = "line " + strconv.Itoa(line) + ":" + strconv.Itoa(column) + " " + msg
}

func (q *qasmErrorListener) ReportAmbiguity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, exact bool, ambigAlts *antlr.BitSet, configs antlr.ATNConfigSet) {
	q.msgBuffer = "reportAmbiguity"
}

func (q *qasmErrorListener) ReportAttemptingFullContext(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, conflictingAlts *antlr.BitSet, configs antlr.ATNConfigSet) {
	q.msgBuffer = "reportAttemptingFullContext"
}

func (q *qasmErrorListener) ReportContextSensitivity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex, prediction int, configs antlr.ATNConfigSet) {
	q.msgBuffer = "reportContextSensitivity"
}

func validateStatements(circ *Circuit, qasmSupport *QASMSupport) error {
	if qasmSupport.AllowList.Enabled {
		if err := filterList(circ, qasmSupport.AllowList.Statements, false); err != nil {
			zap.L().Info(fmt.Sprintf("[AllowList Error] %s", err.Error()))
			return err
		}
	}
	if qasmSupport.DenyList.Enabled {
		if err := filterList(circ, qasmSupport.DenyList.Statements, true); err != nil {
			zap.L().Info(fmt.Sprintf("[DenyList Error] %s", err.Error()))
			return err
		}
	}
	return nil
}

func filterList(circ *Circuit, list []*QASMStatementType, returnIfFiltered bool) error {
	errFunc := func(statement string) error {
		return fmt.Errorf("statement:%s is not supported", statement)
	}
	pc := circ.programContext

	statementList := []string{}
	for _, qt := range list {
		statementList = append(statementList, qt.Name)
	}
	for _, statement := range pc.AllStatement() {
		n := antlr.TreesGetNodeText(statement.GetChild(0), circ.ruleNames, pc.GetParser())
		//usedStatement := &QASMStatementType{Name: n}
		// TODO optimize
		if returnIfFiltered {
			// DenyList
			if common.ContainsStatementName(n, statementList) {
				return errFunc(n)
			}
		} else {
			// AllowList
			if !common.ContainsStatementName(n, statementList) {
				return errFunc(n)
			}
		}
	}
	return nil
}

func checkResource(circ *Circuit, qubinNumber int) error {
	circIR, err := NewCircuitIR(circ.programContext)
	if err != nil {
		return err
	}
	if circIR.ProgramIR.QubitCount > qubinNumber {
		return fmt.Errorf("Too many quibits in your circuit. We only have %d qubits.", qubinNumber)
	}
	return nil
}
