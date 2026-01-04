# BASIC Interpreter

A BASIC language interpreter written in Go.

## Features

- Classic BASIC syntax with line numbers
- Supported statements:
  - `PRINT` - Output text and expressions
  - `LET` - Variable assignment
  - `IF...THEN...ELSE` - Conditional execution
  - `FOR...TO...STEP...NEXT` - Loops
  - `GOTO` - Jump to line number
  - `GOSUB`/`RETURN` - Subroutines
  - `INPUT` - User input
  - `DIM` - Array declaration
  - `REM` - Comments
  - `END` - End program
- Operators: `+`, `-`, `*`, `/`, `MOD`, `<`, `>`, `<=`, `>=`, `==`, `<>`, `AND`, `OR`, `NOT`
- Data types: Numbers and Strings
- Arrays with indexing

## Building

```bash
go build -o basic
```

## Usage

### Run a BASIC file:
```bash
./basic examples/hello.bas
```

### Compile a BASIC file to Go
```bash
./basic -compile hello.go examples/hello.bas
go run hello.go
```

### Interactive REPL:
```bash
./basic
```

REPL commands:
- `RUN` - Execute the program
- `LIST` - Show the program
- `CLEAR` or `NEW` - Clear the program
- `EXIT` or `QUIT` - Exit the interpreter

## Examples

### Hello World
```basic
10 PRINT "Hello, World!"
20 END
```

### For Loop
```basic
10 FOR I = 1 TO 10
20 PRINT I
30 NEXT I
40 END
```

### Input and Conditionals
```basic
10 INPUT "Enter your age: "; AGE
20 IF AGE >= 18 THEN PRINT "Adult" ELSE PRINT "Minor"
30 END
```

## Example Programs

See the `examples/` directory for more sample programs:
- `hello.bas` - Hello World
- `fibonacci.bas` - Fibonacci sequence
- `guess.bas` - Number guessing game
- `multiply.bas` - Multiplication table
- `gosub.bas` - Subroutine example
