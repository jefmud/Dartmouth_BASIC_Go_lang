package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/basis-ex/ast"
	"github.com/basis-ex/compiler"
	"github.com/basis-ex/evaluator"
	"github.com/basis-ex/lexer"
	"github.com/basis-ex/parser"
	"os"
	"sort"
	"strconv"
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

		if upperLine == "DELETE" || strings.HasPrefix(upperLine, "DELETE ") {
			arg := strings.TrimSpace(line[len("DELETE"):])
			if arg == "" {
				fmt.Println("Usage: DELETE <n> or DELETE <n-m>")
				continue
			}
			deleted, err := deleteLines(lines, arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				continue
			}
			if deleted == 0 {
				fmt.Println("No matching lines to delete")
			} else {
				fmt.Printf("Deleted %d line(s)\n", deleted)
			}
			continue
		}

		if upperLine == "LOAD" || strings.HasPrefix(upperLine, "LOAD ") {
			filename := strings.TrimSpace(line[len("LOAD"):])
			if filename == "" {
				fmt.Println("Usage: LOAD <file.bas>")
				continue
			}
			loaded, err := loadProgramFromFile(filename)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading program: %v\n", err)
				continue
			}
			lines = loaded
			fmt.Printf("Loaded %d lines from %s\n", len(lines), filename)
			continue
		}

		if upperLine == "SAVE" || strings.HasPrefix(upperLine, "SAVE ") {
			filename := strings.TrimSpace(line[len("SAVE"):])
			if filename == "" {
				fmt.Println("Usage: SAVE <file.bas>")
				continue
			}
			if len(lines) == 0 {
				fmt.Println("No program to save")
				continue
			}
			if err := saveProgramToFile(lines, filename); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving program: %v\n", err)
				continue
			}
			fmt.Printf("Saved %d lines to %s\n", len(lines), filename)
			continue
		}

		if upperLine == "LIST" || strings.HasPrefix(upperLine, "LIST ") {
			arg := ""
			if len(line) > len("LIST") {
				arg = strings.TrimSpace(line[len("LIST"):])
			}
			if err := listProgram(lines, arg); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
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

		if err := handleProgramInput(program, p.Errors(), line, lines, true, true); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}
	}
}

func runProgram(lines map[int]string) {
	if len(lines) == 0 {
		fmt.Println("No program to run")
		return
	}

	lineNums := sortedLineNumbers(lines)
	var programText strings.Builder
	for _, num := range lineNums {
		programText.WriteString(lines[num])
		programText.WriteByte('\n')
	}

	l := lexer.New(programText.String())
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

func listProgram(lines map[int]string, arg string) error {
	if len(lines) == 0 {
		fmt.Println("No program")
		return nil
	}

	start, end, hasRange, err := parseListArgs(arg)
	if err != nil {
		return err
	}

	lineNums := sortedLineNumbers(lines)
	printed := false
	for _, num := range lineNums {
		if !hasRange || (num >= start && (end == -1 || num <= end)) {
			fmt.Println(lines[num])
			printed = true
		}
	}

	if hasRange && !printed {
		fmt.Println("No matching lines")
	}

	return nil
}

func parseListArgs(arg string) (int, int, bool, error) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return 0, 0, false, nil
	}

	if strings.Contains(arg, "-") {
		parts := strings.SplitN(arg, "-", 2)
		startStr := strings.TrimSpace(parts[0])
		endStr := strings.TrimSpace(parts[1])

		if startStr == "" {
			return 0, 0, false, fmt.Errorf("LIST requires a starting line number")
		}

		start, err := strconv.Atoi(startStr)
		if err != nil {
			return 0, 0, false, fmt.Errorf("invalid start line: %v", err)
		}

		if endStr == "" {
			return start, -1, true, nil
		}

		end, err := strconv.Atoi(endStr)
		if err != nil {
			return 0, 0, false, fmt.Errorf("invalid end line: %v", err)
		}

		if end < start {
			return 0, 0, false, fmt.Errorf("end line must be >= start line")
		}

		return start, end, true, nil
	}

	start, err := strconv.Atoi(arg)
	if err != nil {
		return 0, 0, false, fmt.Errorf("invalid line number: %v", err)
	}

	return start, start, true, nil
}

func handleProgramInput(program *ast.Program, parseErrors []string, rawLine string, lines map[int]string, allowImmediate bool, echoStored bool) error {
	if len(parseErrors) > 0 {
		return fmt.Errorf(strings.Join(parseErrors, "; "))
	}

	if len(program.Statements) == 0 {
		return fmt.Errorf("no statements parsed from input")
	}

	for lineNum := range program.Statements {
		if lineNum > 0 {
			lines[lineNum] = rawLine
			if echoStored {
				fmt.Printf("Line %d stored\n", lineNum)
			}
		} else {
			if !allowImmediate {
				return fmt.Errorf("line must start with a line number")
			}
			eval := evaluator.New(program)
			if err := eval.Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

func loadProgramFromFile(filename string) (map[int]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	loaded := make(map[int]string)
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()
		if err := handleProgramInput(program, p.Errors(), line, loaded, false, false); err != nil {
			return nil, fmt.Errorf("line %q: %w", line, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return loaded, nil
}

func saveProgramToFile(lines map[int]string, filename string) error {
	lineNums := sortedLineNumbers(lines)
	var builder strings.Builder

	for _, num := range lineNums {
		builder.WriteString(lines[num])
		builder.WriteByte('\n')
	}

	return os.WriteFile(filename, []byte(builder.String()), 0644)
}

func sortedLineNumbers(lines map[int]string) []int {
	lineNums := make([]int, 0, len(lines))
	for num := range lines {
		lineNums = append(lineNums, num)
	}
	sort.Ints(lineNums)
	return lineNums
}

func deleteLines(lines map[int]string, arg string) (int, error) {
	start, end, err := parseDeleteArgs(arg)
	if err != nil {
		return 0, err
	}

	deleted := 0
	for num := range lines {
		if num >= start && num <= end {
			delete(lines, num)
			deleted++
		}
	}

	return deleted, nil
}

func parseDeleteArgs(arg string) (int, int, error) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return 0, 0, fmt.Errorf("missing line number")
	}

	if strings.Contains(arg, "-") {
		parts := strings.SplitN(arg, "-", 2)
		startStr := strings.TrimSpace(parts[0])
		endStr := strings.TrimSpace(parts[1])

		if startStr == "" || endStr == "" {
			return 0, 0, fmt.Errorf("DELETE range requires both start and end, e.g. DELETE 10-20")
		}

		start, err := strconv.Atoi(startStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid start line: %v", err)
		}
		end, err := strconv.Atoi(endStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid end line: %v", err)
		}
		if end < start {
			return 0, 0, fmt.Errorf("end line must be >= start line")
		}
		return start, end, nil
	}

	num, err := strconv.Atoi(arg)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid line number: %v", err)
	}
	return num, num, nil
}
