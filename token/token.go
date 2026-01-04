package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"
	NEWLINE = "NEWLINE"

	IDENT  = "IDENT"
	NUMBER = "NUMBER"
	STRING = "STRING"

	ASSIGN = "="
	PLUS   = "+"
	MINUS  = "-"
	MULT   = "*"
	DIV    = "/"
	MOD    = "MOD"

	LT     = "<"
	GT     = ">"
	LE     = "<="
	GE     = ">="
	EQ     = "=="
	NE     = "<>"

	LPAREN = "("
	RPAREN = ")"
	COMMA  = ","
	COLON  = ":"
	SEMICOLON = ";"

	PRINT  = "PRINT"
	LET    = "LET"
	IF     = "IF"
	THEN   = "THEN"
	ELSE   = "ELSE"
	GOTO   = "GOTO"
	GOSUB  = "GOSUB"
	RETURN = "RETURN"
	FOR    = "FOR"
	TO     = "TO"
	STEP   = "STEP"
	NEXT   = "NEXT"
	INPUT  = "INPUT"
	REM    = "REM"
	END    = "END"
	DIM    = "DIM"
	AND    = "AND"
	OR     = "OR"
	NOT    = "NOT"
)

var keywords = map[string]TokenType{
	"PRINT":  PRINT,
	"LET":    LET,
	"IF":     IF,
	"THEN":   THEN,
	"ELSE":   ELSE,
	"GOTO":   GOTO,
	"GOSUB":  GOSUB,
	"RETURN": RETURN,
	"FOR":    FOR,
	"TO":     TO,
	"STEP":   STEP,
	"NEXT":   NEXT,
	"INPUT":  INPUT,
	"REM":    REM,
	"END":    END,
	"DIM":    DIM,
	"AND":    AND,
	"OR":     OR,
	"NOT":    NOT,
	"MOD":    MOD,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
