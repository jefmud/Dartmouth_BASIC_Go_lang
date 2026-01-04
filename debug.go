package main

import (
	"fmt"
	"github.com/basis-ex/evaluator"
	"github.com/basis-ex/lexer"
	"github.com/basis-ex/parser"
)

func main() {
	code := `10 REM Hello World Program
20 PRINT "Hello, World!"
30 END`

	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Println("\t" + msg)
		}
		return
	}

	fmt.Printf("Program has %d statements\n", len(program.Statements))
	for lineNum, stmt := range program.Statements {
		fmt.Printf("Line %d: %T\n", lineNum, stmt)
	}

	eval := evaluator.New(program)
	if err := eval.Run(); err != nil {
		fmt.Printf("Runtime error: %v\n", err)
	}
}
