package compiler

import (
	"fmt"
	"sort"
	"strings"

	"github.com/basis-ex/ast"
)

// Compile converts a parsed BASIC program into a standalone Go source file.
func Compile(program *ast.Program) (string, error) {
	lines := make([]int, 0, len(program.Statements))
	for line := range program.Statements {
		lines = append(lines, line)
	}
	sort.Ints(lines)

	lineIndex := make(map[int]int, len(lines))
	for i, line := range lines {
		lineIndex[line] = i
	}

	var out strings.Builder

	out.WriteString("package main\n\n")
	out.WriteString("import (\n")
	out.WriteString("\t\"bufio\"\n\t\"fmt\"\n\t\"math\"\n\t\"os\"\n\t\"strconv\"\n\t\"strings\"\n")
	out.WriteString(")\n\n")
	out.WriteString("// keep imports used even for tiny programs\n")
	out.WriteString("var _ = []interface{}{strconv.ParseFloat, strings.TrimSpace}\n\n")
	out.WriteString(runtimeHelpers)

	fmt.Fprintf(&out, "var programLines = []int{%s}\n", joinInts(lines, ","))
	out.WriteString("var lineIndex = map[int]int{\n")
	for _, line := range lines {
		fmt.Fprintf(&out, "\t%d: %d,\n", line, lineIndex[line])
	}
	out.WriteString("}\n\n")

	out.WriteString("func run() error {\n")
	out.WriteString("\tenv := newEnv()\n")
	out.WriteString("\tcallStack := []int{}\n")
	out.WriteString("\tforLoops := map[string]*forLoopState{}\n")
	out.WriteString("\thalted := false\n")
	out.WriteString("\tpc := 0\n")
	out.WriteString("\t_ = env\n\t_ = callStack\n\t_ = forLoops\n\n")
	out.WriteString("\tfor pc < len(programLines) && !halted {\n")
	out.WriteString("\t\tswitch programLines[pc] {\n")

	tmpCounter := 0
	for _, line := range lines {
		stmt := program.Statements[line]
		out.WriteString(fmt.Sprintf("\t\tcase %d:\n", line))
		out.WriteString("\t\t\t{\n")
		emitter := newEmitter(&out, "\t\t\t\t", &tmpCounter)
		if err := emitStatement(emitter, stmt); err != nil {
			return "", err
		}
		out.WriteString("\t\t\t}\n")
	}

	out.WriteString("\t\tdefault:\n")
	out.WriteString("\t\t\treturn fmt.Errorf(\"unknown line %d\", programLines[pc])\n")
	out.WriteString("\t\t}\n")
	out.WriteString("\t\tpc++\n")
	out.WriteString("\t}\n")
	out.WriteString("\treturn nil\n")
	out.WriteString("}\n\n")

	out.WriteString("func main() {\n")
	out.WriteString("\tif err := run(); err != nil {\n")
	out.WriteString("\t\tfmt.Fprintf(os.Stderr, \"error: %v\\n\", err)\n")
	out.WriteString("\t\tos.Exit(1)\n")
	out.WriteString("\t}\n")
	out.WriteString("}\n")

	return out.String(), nil
}

// emitter helps build Go code while keeping indentation and unique temp names.
type emitter struct {
	buf     *strings.Builder
	indent  string
	counter *int
}

func newEmitter(buf *strings.Builder, indent string, counter *int) *emitter {
	return &emitter{buf: buf, indent: indent, counter: counter}
}

func (e *emitter) line(format string, args ...interface{}) {
	fmt.Fprintf(e.buf, "%s%s\n", e.indent, fmt.Sprintf(format, args...))
}

func (e *emitter) temp() string {
	*e.counter++
	return fmt.Sprintf("tmp%d", *e.counter)
}

func (e *emitter) nested() *emitter {
	return &emitter{buf: e.buf, indent: e.indent + "\t", counter: e.counter}
}

func emitStatement(e *emitter, stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.PrintStatement:
		return emitPrint(e, s)
	case *ast.LetStatement:
		return emitLet(e, s)
	case *ast.IfStatement:
		return emitIf(e, s)
	case *ast.GotoStatement:
		return emitGoto(e, s)
	case *ast.GosubStatement:
		return emitGosub(e, s)
	case *ast.ReturnStatement:
		e.line("if len(callStack) == 0 {")
		e.nested().line("return fmt.Errorf(\"RETURN without GOSUB\")")
		e.line("}")
		e.line("pc = callStack[len(callStack)-1]")
		e.line("callStack = callStack[:len(callStack)-1]")
		return nil
	case *ast.ForStatement:
		return emitFor(e, s)
	case *ast.NextStatement:
		return emitNext(e, s)
	case *ast.InputStatement:
		return emitInput(e, s)
	case *ast.EndStatement:
		e.line("halted = true")
		return nil
	case *ast.RemStatement:
		return nil
	case *ast.DimStatement:
		e.line("env.ensureArray(%q)", s.Name.Value)
		return nil
	case *ast.ExpressionStatement:
		val, err := emitExpression(e, s.Expression)
		if err != nil {
			return err
		}
		e.line("_ = %s", val)
		return nil
	case *ast.SequenceStatement:
		for _, inner := range s.Statements {
			if err := emitStatement(e, inner); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("compiler: unsupported statement %T", stmt)
	}
}

func emitPrint(e *emitter, stmt *ast.PrintStatement) error {
	if len(stmt.Expressions) == 0 {
		e.line("fmt.Println()")
		return nil
	}

	for i, expr := range stmt.Expressions {
		val, err := emitExpression(e, expr)
		if err != nil {
			return err
		}
		e.line("fmt.Print(%s.inspect())", val)

		if i < len(stmt.Separators) {
			sep := stmt.Separators[i]
			e.line("fmt.Print(%q)", sep)
		}
	}

	if stmt.TrailingNewline {
		e.line("fmt.Println()")
	}
	return nil
}

func emitLet(e *emitter, stmt *ast.LetStatement) error {
	val, err := emitExpression(e, stmt.Value)
	if err != nil {
		return err
	}
	e.line("env.set(%q, %s)", stmt.Name.Value, val)
	return nil
}

func emitIf(e *emitter, stmt *ast.IfStatement) error {
	cond, err := emitExpression(e, stmt.Condition)
	if err != nil {
		return err
	}
	e.line("if truthy(%s) {", cond)
	if err := emitStatement(e.nested(), stmt.Consequence); err != nil {
		return err
	}
	if stmt.Alternative != nil {
		e.line("} else {")
		if err := emitStatement(e.nested(), stmt.Alternative); err != nil {
			return err
		}
	}
	e.line("}")
	return nil
}

func emitGoto(e *emitter, stmt *ast.GotoStatement) error {
	targetVal, err := emitExpression(e, stmt.LineNumber)
	if err != nil {
		return err
	}
	numVar := e.temp()
	e.line("%s, err := mustNumber(%s)", numVar, targetVal)
	e.line("if err != nil {")
	e.nested().line("return fmt.Errorf(\"GOTO requires a number\")")
	e.line("}")
	e.line("lineNum := int(%s)", numVar)
	e.line("idx, ok := lineIndex[lineNum]")
	e.line("if !ok {")
	e.nested().line("return fmt.Errorf(\"line %d not found\", lineNum)")
	e.line("}")
	e.line("pc = idx - 1")
	return nil
}

func emitGosub(e *emitter, stmt *ast.GosubStatement) error {
	targetVal, err := emitExpression(e, stmt.LineNumber)
	if err != nil {
		return err
	}
	numVar := e.temp()
	e.line("%s, err := mustNumber(%s)", numVar, targetVal)
	e.line("if err != nil {")
	e.nested().line("return fmt.Errorf(\"GOSUB requires a number\")")
	e.line("}")
	e.line("lineNum := int(%s)", numVar)
	e.line("idx, ok := lineIndex[lineNum]")
	e.line("if !ok {")
	e.nested().line("return fmt.Errorf(\"line %d not found\", lineNum)")
	e.line("}")
	e.line("callStack = append(callStack, pc)")
	e.line("pc = idx - 1")
	return nil
}

func emitFor(e *emitter, stmt *ast.ForStatement) error {
	startVal, err := emitExpression(e, stmt.Start)
	if err != nil {
		return err
	}
	endVal, err := emitExpression(e, stmt.End)
	if err != nil {
		return err
	}
	stepVal, err := emitExpression(e, stmt.Step)
	if err != nil {
		return err
	}

	startNum := e.temp()
	endNum := e.temp()
	stepNum := e.temp()

	e.line("%s, err := mustNumber(%s)", startNum, startVal)
	e.line("if err != nil {")
	e.nested().line("return fmt.Errorf(\"FOR start value must be a number\")")
	e.line("}")
	e.line("%s, err := mustNumber(%s)", endNum, endVal)
	e.line("if err != nil {")
	e.nested().line("return fmt.Errorf(\"FOR end value must be a number\")")
	e.line("}")
	e.line("%s, err := mustNumber(%s)", stepNum, stepVal)
	e.line("if err != nil {")
	e.nested().line("return fmt.Errorf(\"FOR step value must be a number\")")
	e.line("}")

	e.line("env.set(%q, numVal(%s))", stmt.Variable.Value, startNum)
	e.line("forLoops[%q] = &forLoopState{End: %s, Step: %s, StartPC: pc}", stmt.Variable.Value, endNum, stepNum)
	return nil
}

func emitNext(e *emitter, stmt *ast.NextStatement) error {
	varName := stmt.Variable
	if varName != nil {
		e.line("loopName := %q", varName.Value)
	} else {
		e.line("loopName := \"\"")
		e.line("for name := range forLoops {")
		e.nested().line("loopName = name")
		e.nested().line("break")
		e.line("}")
	}

	e.line("if loopName == \"\" {")
	e.nested().line("return fmt.Errorf(\"NEXT without FOR\")")
	e.line("}")

	e.line("loopState, ok := forLoops[loopName]")
	e.line("if !ok {")
	e.nested().line("return fmt.Errorf(\"NEXT without matching FOR\")")
	e.line("}")

	e.line("val := env.get(loopName)")
	e.line("if !val.isNumber() {")
	e.nested().line("return fmt.Errorf(\"loop variable must be a number\")")
	e.line("}")

	newVal := e.temp()
	e.line("%s := val.num + loopState.Step", newVal)
	e.line("shouldContinue := false")
	e.line("if loopState.Step > 0 {")
	e.nested().line("shouldContinue = %s <= loopState.End", newVal)
	e.line("} else {")
	e.nested().line("shouldContinue = %s >= loopState.End", newVal)
	e.line("}")

	e.line("if shouldContinue {")
	e.nested().line("env.set(loopName, numVal(%s))", newVal)
	e.nested().line("pc = loopState.StartPC")
	e.line("} else {")
	e.nested().line("delete(forLoops, loopName)")
	e.line("}")
	return nil
}

func emitInput(e *emitter, stmt *ast.InputStatement) error {
	if stmt.Prompt != "" {
		prompt := stmt.Prompt
		e.line("fmt.Print(%q)", prompt)
		if !strings.HasSuffix(prompt, " ") {
			e.line("fmt.Print(\" \")")
		}
	}

	e.line("line, err := env.reader.ReadString('\\n')")
	e.line("if err != nil {")
	e.nested().line("return err")
	e.line("}")
	e.line("line = strings.TrimSpace(line)")
	e.line("parts := strings.Split(line, \",\")")

	for i, ident := range stmt.Variables {
		valVar := e.temp()
		e.line("var %s Value", valVar)
		e.line("if len(parts) > %d {", i)
		valEmitter := e.nested()
		valEmitter.line("text := strings.TrimSpace(parts[%d])", i)
		valEmitter.line("if num, err := strconv.ParseFloat(text, 64); err == nil {")
		valEmitter.nested().line("%s = numVal(num)", valVar)
		valEmitter.line("} else {")
		valEmitter.nested().line("%s = strVal(text)", valVar)
		valEmitter.line("}")
		e.line("} else {")
		e.nested().line("%s = numVal(0)", valVar)
		e.line("}")
		e.line("env.set(%q, %s)", ident.Value, valVar)
	}
	return nil
}

func emitExpression(e *emitter, expr ast.Expression) (string, error) {
	switch node := expr.(type) {
	case *ast.NumberLiteral:
		tmp := e.temp()
		e.line("%s := numVal(%g)", tmp, node.Value)
		return tmp, nil
	case *ast.StringLiteral:
		tmp := e.temp()
		e.line("%s := strVal(%q)", tmp, node.Value)
		return tmp, nil
	case *ast.Identifier:
		tmp := e.temp()
		e.line("%s := env.get(%q)", tmp, node.Value)
		return tmp, nil
	case *ast.InfixExpression:
		left, err := emitExpression(e, node.Left)
		if err != nil {
			return "", err
		}
		right, err := emitExpression(e, node.Right)
		if err != nil {
			return "", err
		}
		tmp := e.temp()
		e.line("%s, err := applyInfix(%q, %s, %s)", tmp, node.Operator, left, right)
		e.line("if err != nil {")
		e.nested().line("return err")
		e.line("}")
		return tmp, nil
	case *ast.PrefixExpression:
		right, err := emitExpression(e, node.Right)
		if err != nil {
			return "", err
		}
		tmp := e.temp()
		e.line("%s, err := applyPrefix(%q, %s)", tmp, node.Operator, right)
		e.line("if err != nil {")
		e.nested().line("return err")
		e.line("}")
		return tmp, nil
	case *ast.ArrayAccess:
		index, err := emitExpression(e, node.Index)
		if err != nil {
			return "", err
		}
		tmp := e.temp()
		e.line("%s, err := arrayAccess(env, %q, %s)", tmp, node.Name.Value, index)
		e.line("if err != nil {")
		e.nested().line("return err")
		e.line("}")
		return tmp, nil
	default:
		return "", fmt.Errorf("compiler: unsupported expression %T", expr)
	}
}

func joinInts(values []int, sep string) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(parts, sep)
}

const runtimeHelpers = `
type valueKind int

const (
	numberKind valueKind = iota
	stringKind
)

type Value struct {
	kind valueKind
	num  float64
	str  string
}

func numVal(v float64) Value { return Value{kind: numberKind, num: v} }
func strVal(v string) Value  { return Value{kind: stringKind, str: v} }

func (v Value) isNumber() bool { return v.kind == numberKind }
func (v Value) inspect() string {
	if v.kind == numberKind {
		return fmt.Sprintf("%g", v.num)
	}
	return v.str
}

type env struct {
	vars   map[string]Value
	arrays map[string]map[int]Value
	reader *bufio.Reader
}

func newEnv() *env {
	return &env{
		vars:   map[string]Value{},
		arrays: map[string]map[int]Value{},
		reader: bufio.NewReader(os.Stdin),
	}
}

func (e *env) get(name string) Value {
	if v, ok := e.vars[name]; ok {
		return v
	}
	return numVal(0)
}

func (e *env) set(name string, val Value) {
	e.vars[name] = val
}

func (e *env) ensureArray(name string) {
	if _, ok := e.arrays[name]; !ok {
		e.arrays[name] = map[int]Value{}
	}
}

func (e *env) array(name string) (map[int]Value, bool) {
	arr, ok := e.arrays[name]
	return arr, ok
}

type forLoopState struct {
	End     float64
	Step    float64
	StartPC int
}

func mustNumber(v Value) (float64, error) {
	if !v.isNumber() {
		return 0, fmt.Errorf("expected number")
	}
	return v.num, nil
}

func truthy(v Value) bool {
	if v.kind == numberKind {
		return v.num != 0
	}
	return v.str != ""
}

func applyPrefix(op string, right Value) (Value, error) {
	switch op {
	case "-":
		if !right.isNumber() {
			return Value{}, fmt.Errorf("cannot negate non-number")
		}
		return numVal(-right.num), nil
	case "NOT":
		if truthy(right) {
			return numVal(0), nil
		}
		return numVal(1), nil
	default:
		return Value{}, fmt.Errorf("unknown operator: %s", op)
	}
}

func applyInfix(op string, left, right Value) (Value, error) {
	if left.isNumber() && right.isNumber() {
		switch op {
		case "+":
			return numVal(left.num + right.num), nil
		case "-":
			return numVal(left.num - right.num), nil
		case "*":
			return numVal(left.num * right.num), nil
		case "/":
			if right.num == 0 {
				return Value{}, fmt.Errorf("division by zero")
			}
			return numVal(left.num / right.num), nil
		case "MOD":
			return numVal(math.Mod(left.num, right.num)), nil
		case "<":
			if left.num < right.num {
				return numVal(1), nil
			}
			return numVal(0), nil
		case ">":
			if left.num > right.num {
				return numVal(1), nil
			}
			return numVal(0), nil
		case "<=":
			if left.num <= right.num {
				return numVal(1), nil
			}
			return numVal(0), nil
		case ">=":
			if left.num >= right.num {
				return numVal(1), nil
			}
			return numVal(0), nil
		case "==":
			if left.num == right.num {
				return numVal(1), nil
			}
			return numVal(0), nil
		case "<>":
			if left.num != right.num {
				return numVal(1), nil
			}
			return numVal(0), nil
		case "AND":
			if truthy(left) && truthy(right) {
				return numVal(1), nil
			}
			return numVal(0), nil
		case "OR":
			if truthy(left) || truthy(right) {
				return numVal(1), nil
			}
			return numVal(0), nil
		}
	}

	if left.kind == stringKind && right.kind == stringKind {
		switch op {
		case "+":
			return strVal(left.str + right.str), nil
		case "==":
			if left.str == right.str {
				return numVal(1), nil
			}
			return numVal(0), nil
		case "<>":
			if left.str != right.str {
				return numVal(1), nil
			}
			return numVal(0), nil
		}
	}

	return Value{}, fmt.Errorf("unsupported operation: %s %s %s", left.inspect(), op, right.inspect())
}

func arrayAccess(env *env, name string, index Value) (Value, error) {
	arr, ok := env.array(name)
	if !ok {
		return Value{}, fmt.Errorf("array %s not defined", name)
	}

	idx, err := mustNumber(index)
	if err != nil {
		return Value{}, fmt.Errorf("array index must be a number")
	}

	val, ok := arr[int(idx)]
	if !ok {
		return numVal(0), nil
	}

	return val, nil
}
`
