package evaluator

import (
	"bufio"
	"fmt"
	"github.com/basis-ex/ast"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

type ValueType string

const (
	NUMBER_VAL ValueType = "NUMBER"
	STRING_VAL ValueType = "STRING"
	ARRAY_VAL  ValueType = "ARRAY"
)

type Value interface {
	Type() ValueType
	Inspect() string
}

type NumberValue struct {
	Value float64
}

func (n *NumberValue) Type() ValueType { return NUMBER_VAL }
func (n *NumberValue) Inspect() string { return fmt.Sprintf("%g", n.Value) }

type StringValue struct {
	Value string
}

func (s *StringValue) Type() ValueType { return STRING_VAL }
func (s *StringValue) Inspect() string { return s.Value }

type ArrayValue struct {
	Elements map[int]Value
}

func (a *ArrayValue) Type() ValueType { return ARRAY_VAL }
func (a *ArrayValue) Inspect() string { return "[ARRAY]" }

type Environment struct {
	variables map[string]Value
	arrays    map[string]*ArrayValue
	reader    *bufio.Reader
}

func NewEnvironment() *Environment {
	return &Environment{
		variables: make(map[string]Value),
		arrays:    make(map[string]*ArrayValue),
		reader:    bufio.NewReader(os.Stdin),
	}
}

func (e *Environment) Get(name string) (Value, bool) {
	val, ok := e.variables[name]
	return val, ok
}

func (e *Environment) Set(name string, val Value) {
	e.variables[name] = val
}

func (e *Environment) GetArray(name string) (*ArrayValue, bool) {
	arr, ok := e.arrays[name]
	return arr, ok
}

func (e *Environment) SetArray(name string, arr *ArrayValue) {
	e.arrays[name] = arr
}

type Evaluator struct {
	env         *Environment
	program     *ast.Program
	lines       []int
	currentLine int
	callStack   []int
	forLoops    map[string]*ForLoopState
	halted      bool
}

type ForLoopState struct {
	Variable  string
	End       float64
	Step      float64
	NextLine  int
	StartLine int
}

func New(program *ast.Program) *Evaluator {
	lines := make([]int, 0, len(program.Statements))
	for lineNum := range program.Statements {
		lines = append(lines, lineNum)
	}
	sort.Ints(lines)

	return &Evaluator{
		env:       NewEnvironment(),
		program:   program,
		lines:     lines,
		callStack: []int{},
		forLoops:  make(map[string]*ForLoopState),
		halted:    false,
	}
}

func (e *Evaluator) Run() error {
	if len(e.lines) == 0 {
		return nil
	}

	e.currentLine = 0

	for e.currentLine < len(e.lines) && !e.halted {
		lineNum := e.lines[e.currentLine]
		stmt := e.program.Statements[lineNum]

		err := e.evalStatement(stmt)
		if err != nil {
			return fmt.Errorf("error at line %d: %v", lineNum, err)
		}

		e.currentLine++
	}

	return nil
}

func (e *Evaluator) evalStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.PrintStatement:
		return e.evalPrintStatement(s)
	case *ast.LetStatement:
		return e.evalLetStatement(s)
	case *ast.IfStatement:
		return e.evalIfStatement(s)
	case *ast.GotoStatement:
		return e.evalGotoStatement(s)
	case *ast.GosubStatement:
		return e.evalGosubStatement(s)
	case *ast.ReturnStatement:
		return e.evalReturnStatement(s)
	case *ast.ForStatement:
		return e.evalForStatement(s)
	case *ast.NextStatement:
		return e.evalNextStatement(s)
	case *ast.InputStatement:
		return e.evalInputStatement(s)
	case *ast.EndStatement:
		e.halted = true
		return nil
	case *ast.RemStatement:
		return nil
	case *ast.DimStatement:
		return e.evalDimStatement(s)
	case *ast.ExpressionStatement:
		_, err := e.evalExpression(s.Expression)
		return err
	case *ast.SequenceStatement:
		for _, inner := range s.Statements {
			if err := e.evalStatement(inner); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown statement type: %T", stmt)
	}
}

func (e *Evaluator) evalPrintStatement(stmt *ast.PrintStatement) error {
	if len(stmt.Expressions) == 0 {
		fmt.Println()
		return nil
	}

	for i, expr := range stmt.Expressions {
		val, err := e.evalExpression(expr)
		if err != nil {
			return err
		}

		fmt.Print(val.Inspect())

		if i < len(stmt.Separators) {
			fmt.Print(stmt.Separators[i])
		}
	}

	if stmt.TrailingNewline {
		fmt.Println()
	}

	return nil
}

func (e *Evaluator) evalLetStatement(stmt *ast.LetStatement) error {
	val, err := e.evalExpression(stmt.Value)
	if err != nil {
		return err
	}

	e.env.Set(stmt.Name.Value, val)
	return nil
}

func (e *Evaluator) evalIfStatement(stmt *ast.IfStatement) error {
	condition, err := e.evalExpression(stmt.Condition)
	if err != nil {
		return err
	}

	if isTruthy(condition) {
		return e.evalStatement(stmt.Consequence)
	} else if stmt.Alternative != nil {
		return e.evalStatement(stmt.Alternative)
	}

	return nil
}

func (e *Evaluator) evalGotoStatement(stmt *ast.GotoStatement) error {
	lineVal, err := e.evalExpression(stmt.LineNumber)
	if err != nil {
		return err
	}

	numVal, ok := lineVal.(*NumberValue)
	if !ok {
		return fmt.Errorf("GOTO requires a number")
	}

	targetLine := int(numVal.Value)
	for i, line := range e.lines {
		if line == targetLine {
			e.currentLine = i - 1
			return nil
		}
	}

	return fmt.Errorf("line %d not found", targetLine)
}

func (e *Evaluator) evalGosubStatement(stmt *ast.GosubStatement) error {
	lineVal, err := e.evalExpression(stmt.LineNumber)
	if err != nil {
		return err
	}

	numVal, ok := lineVal.(*NumberValue)
	if !ok {
		return fmt.Errorf("GOSUB requires a number")
	}

	e.callStack = append(e.callStack, e.currentLine)

	targetLine := int(numVal.Value)
	for i, line := range e.lines {
		if line == targetLine {
			e.currentLine = i - 1
			return nil
		}
	}

	return fmt.Errorf("line %d not found", targetLine)
}

func (e *Evaluator) evalReturnStatement(stmt *ast.ReturnStatement) error {
	if len(e.callStack) == 0 {
		return fmt.Errorf("RETURN without GOSUB")
	}

	e.currentLine = e.callStack[len(e.callStack)-1]
	e.callStack = e.callStack[:len(e.callStack)-1]

	return nil
}

func (e *Evaluator) evalForStatement(stmt *ast.ForStatement) error {
	startVal, err := e.evalExpression(stmt.Start)
	if err != nil {
		return err
	}

	startNum, ok := startVal.(*NumberValue)
	if !ok {
		return fmt.Errorf("FOR start value must be a number")
	}

	endVal, err := e.evalExpression(stmt.End)
	if err != nil {
		return err
	}

	endNum, ok := endVal.(*NumberValue)
	if !ok {
		return fmt.Errorf("FOR end value must be a number")
	}

	stepVal, err := e.evalExpression(stmt.Step)
	if err != nil {
		return err
	}

	stepNum, ok := stepVal.(*NumberValue)
	if !ok {
		return fmt.Errorf("FOR step value must be a number")
	}

	e.env.Set(stmt.Variable.Value, startNum)

	e.forLoops[stmt.Variable.Value] = &ForLoopState{
		Variable:  stmt.Variable.Value,
		End:       endNum.Value,
		Step:      stepNum.Value,
		StartLine: e.currentLine,
	}

	return nil
}

func (e *Evaluator) evalNextStatement(stmt *ast.NextStatement) error {
	var varName string
	if stmt.Variable != nil {
		varName = stmt.Variable.Value
	} else {
		for name := range e.forLoops {
			varName = name
			break
		}
	}

	if varName == "" {
		return fmt.Errorf("NEXT without FOR")
	}

	loopState, ok := e.forLoops[varName]
	if !ok {
		return fmt.Errorf("NEXT without matching FOR")
	}

	val, ok := e.env.Get(varName)
	if !ok {
		return fmt.Errorf("loop variable %s not found", varName)
	}

	numVal, ok := val.(*NumberValue)
	if !ok {
		return fmt.Errorf("loop variable must be a number")
	}

	newVal := numVal.Value + loopState.Step

	shouldContinue := false
	if loopState.Step > 0 {
		shouldContinue = newVal <= loopState.End
	} else {
		shouldContinue = newVal >= loopState.End
	}

	if shouldContinue {
		e.env.Set(varName, &NumberValue{Value: newVal})
		e.currentLine = loopState.StartLine
	} else {
		delete(e.forLoops, varName)
	}

	return nil
}

func (e *Evaluator) evalInputStatement(stmt *ast.InputStatement) error {
	if stmt.Prompt != "" {
		fmt.Print(stmt.Prompt)
		if !strings.HasSuffix(stmt.Prompt, " ") {
			fmt.Print(" ")
		}
	}

	input, err := e.env.reader.ReadString('\n')
	if err != nil {
		return err
	}

	input = strings.TrimSpace(input)
	values := strings.Split(input, ",")

	for i, variable := range stmt.Variables {
		if i >= len(values) {
			e.env.Set(variable.Value, &NumberValue{Value: 0})
			continue
		}

		val := strings.TrimSpace(values[i])
		if num, err := strconv.ParseFloat(val, 64); err == nil {
			e.env.Set(variable.Value, &NumberValue{Value: num})
		} else {
			e.env.Set(variable.Value, &StringValue{Value: val})
		}
	}

	return nil
}

func (e *Evaluator) evalDimStatement(stmt *ast.DimStatement) error {
	sizeVal, err := e.evalExpression(stmt.Size)
	if err != nil {
		return err
	}

	_, ok := sizeVal.(*NumberValue)
	if !ok {
		return fmt.Errorf("DIM size must be a number")
	}

	arr := &ArrayValue{Elements: make(map[int]Value)}
	e.env.SetArray(stmt.Name.Value, arr)

	return nil
}

func (e *Evaluator) evalExpression(expr ast.Expression) (Value, error) {
	switch node := expr.(type) {
	case *ast.NumberLiteral:
		return &NumberValue{Value: node.Value}, nil
	case *ast.StringLiteral:
		return &StringValue{Value: node.Value}, nil
	case *ast.Identifier:
		val, ok := e.env.Get(node.Value)
		if !ok {
			return &NumberValue{Value: 0}, nil
		}
		return val, nil
	case *ast.InfixExpression:
		return e.evalInfixExpression(node)
	case *ast.PrefixExpression:
		return e.evalPrefixExpression(node)
	case *ast.ArrayAccess:
		return e.evalArrayAccess(node)
	default:
		return nil, fmt.Errorf("unknown expression type: %T", expr)
	}
}

func (e *Evaluator) evalInfixExpression(expr *ast.InfixExpression) (Value, error) {
	left, err := e.evalExpression(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := e.evalExpression(expr.Right)
	if err != nil {
		return nil, err
	}

	leftNum, leftIsNum := left.(*NumberValue)
	rightNum, rightIsNum := right.(*NumberValue)

	if leftIsNum && rightIsNum {
		switch expr.Operator {
		case "+":
			return &NumberValue{Value: leftNum.Value + rightNum.Value}, nil
		case "-":
			return &NumberValue{Value: leftNum.Value - rightNum.Value}, nil
		case "*":
			return &NumberValue{Value: leftNum.Value * rightNum.Value}, nil
		case "/":
			if rightNum.Value == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return &NumberValue{Value: leftNum.Value / rightNum.Value}, nil
		case "MOD":
			return &NumberValue{Value: math.Mod(leftNum.Value, rightNum.Value)}, nil
		case "<":
			if leftNum.Value < rightNum.Value {
				return &NumberValue{Value: 1}, nil
			}
			return &NumberValue{Value: 0}, nil
		case ">":
			if leftNum.Value > rightNum.Value {
				return &NumberValue{Value: 1}, nil
			}
			return &NumberValue{Value: 0}, nil
		case "<=":
			if leftNum.Value <= rightNum.Value {
				return &NumberValue{Value: 1}, nil
			}
			return &NumberValue{Value: 0}, nil
		case ">=":
			if leftNum.Value >= rightNum.Value {
				return &NumberValue{Value: 1}, nil
			}
			return &NumberValue{Value: 0}, nil
		case "==":
			if leftNum.Value == rightNum.Value {
				return &NumberValue{Value: 1}, nil
			}
			return &NumberValue{Value: 0}, nil
		case "<>":
			if leftNum.Value != rightNum.Value {
				return &NumberValue{Value: 1}, nil
			}
			return &NumberValue{Value: 0}, nil
		case "AND":
			if isTruthy(left) && isTruthy(right) {
				return &NumberValue{Value: 1}, nil
			}
			return &NumberValue{Value: 0}, nil
		case "OR":
			if isTruthy(left) || isTruthy(right) {
				return &NumberValue{Value: 1}, nil
			}
			return &NumberValue{Value: 0}, nil
		}
	}

	leftStr, leftIsStr := left.(*StringValue)
	rightStr, rightIsStr := right.(*StringValue)

	if leftIsStr && rightIsStr {
		switch expr.Operator {
		case "+":
			return &StringValue{Value: leftStr.Value + rightStr.Value}, nil
		case "==":
			if leftStr.Value == rightStr.Value {
				return &NumberValue{Value: 1}, nil
			}
			return &NumberValue{Value: 0}, nil
		case "<>":
			if leftStr.Value != rightStr.Value {
				return &NumberValue{Value: 1}, nil
			}
			return &NumberValue{Value: 0}, nil
		}
	}

	return nil, fmt.Errorf("unsupported operation: %s %s %s", left.Type(), expr.Operator, right.Type())
}

func (e *Evaluator) evalPrefixExpression(expr *ast.PrefixExpression) (Value, error) {
	right, err := e.evalExpression(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case "-":
		if num, ok := right.(*NumberValue); ok {
			return &NumberValue{Value: -num.Value}, nil
		}
		return nil, fmt.Errorf("cannot negate non-number")
	case "NOT":
		if isTruthy(right) {
			return &NumberValue{Value: 0}, nil
		}
		return &NumberValue{Value: 1}, nil
	default:
		return nil, fmt.Errorf("unknown operator: %s", expr.Operator)
	}
}

func (e *Evaluator) evalArrayAccess(expr *ast.ArrayAccess) (Value, error) {
	arr, ok := e.env.GetArray(expr.Name.Value)
	if !ok {
		return nil, fmt.Errorf("array %s not defined", expr.Name.Value)
	}

	indexVal, err := e.evalExpression(expr.Index)
	if err != nil {
		return nil, err
	}

	indexNum, ok := indexVal.(*NumberValue)
	if !ok {
		return nil, fmt.Errorf("array index must be a number")
	}

	index := int(indexNum.Value)
	val, ok := arr.Elements[index]
	if !ok {
		return &NumberValue{Value: 0}, nil
	}

	return val, nil
}

func isTruthy(val Value) bool {
	switch v := val.(type) {
	case *NumberValue:
		return v.Value != 0
	case *StringValue:
		return v.Value != ""
	default:
		return false
	}
}
