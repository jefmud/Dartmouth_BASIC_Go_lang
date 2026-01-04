package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/basis-ex/compiler"
	"github.com/basis-ex/evaluator"
	"github.com/basis-ex/lexer"
	"github.com/basis-ex/parser"
	"os"
	"strings"
)

func main() {
	compileOut := flag.String("compile", "", "write Go source for the BASIC program to this file (use '-' for stdout)")
	flag.Parse()

	args := flag.Args()
	if *compileOut != "" {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "compile mode requires a BASIC file argument")
			os.Exit(1)
		}
		compileFile(args[0], *compileOut)
		return
	}

	if len(args) > 0 {
		runFile(args[0])
		return
	}

	runREPL()
}

func runFile(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Println("\t" + msg)
		}
		os.Exit(1)
	}

	eval := evaluator.New(program)
	if err := eval.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}
}

func compileFile(filename, output string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Println("\t" + msg)
		}
		os.Exit(1)
	}

	code, err := compiler.Compile(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compile error: %v\n", err)
		os.Exit(1)
	}

	if output == "-" {
		fmt.Print(code)
		return
	}

	if err := os.WriteFile(output, []byte(code), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Go source written to %s\n", output)
	fmt.Printf("Build with: go build -o basic_out %s\n", output)
	fmt.Printf("Run with:   go run %s\n", output)
}

func runREPL() {
	fmt.Println("BASIC Interpreter v1.0")
	fmt.Println("Type 'EXIT' to quit, 'RUN' to execute, 'LIST' to show program")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	lines := make(map[int]string)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		upperLine := strings.ToUpper(line)

		if upperLine == "EXIT" || upperLine == "QUIT" {
			break
		}

		if upperLine == "RUN" {
			runProgram(lines)
			continue
		}

		if upperLine == "LIST" {
			listProgram(lines)
			continue
		}

		if upperLine == "CLEAR" || upperLine == "NEW" {
			lines = make(map[int]string)
			fmt.Println("Program cleared")
			continue
		}

		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			fmt.Println("Errors:")
			for _, msg := range p.Errors() {
				fmt.Println("\t" + msg)
			}
			continue
		}

		if len(program.Statements) == 1 {
			for lineNum := range program.Statements {
				if lineNum > 0 {
					lines[lineNum] = line
					fmt.Printf("Line %d stored\n", lineNum)
				} else {
					eval := evaluator.New(program)
					if err := eval.Run(); err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					}
				}
			}
		}
	}
}

func runProgram(lines map[int]string) {
	if len(lines) == 0 {
		fmt.Println("No program to run")
		return
	}

	programText := ""
	for _, line := range lines {
		programText += line + "\n"
	}

	l := lexer.New(programText)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Println("\t" + msg)
		}
		return
	}

	eval := evaluator.New(program)
	if err := eval.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
	}
}

func listProgram(lines map[int]string) {
	if len(lines) == 0 {
		fmt.Println("No program")
		return
	}

	lineNums := make([]int, 0, len(lines))
	for num := range lines {
		lineNums = append(lineNums, num)
	}

	for i := 0; i < len(lineNums); i++ {
		for j := i + 1; j < len(lineNums); j++ {
			if lineNums[i] > lineNums[j] {
				lineNums[i], lineNums[j] = lineNums[j], lineNums[i]
			}
		}
	}

	for _, num := range lineNums {
		fmt.Println(lines[num])
	}
}
