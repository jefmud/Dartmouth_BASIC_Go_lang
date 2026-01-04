package parser

import (
	"fmt"
	"github.com/basis-ex/ast"
	"github.com/basis-ex/lexer"
	"github.com/basis-ex/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	LOGICAL
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

var precedences = map[token.TokenType]int{
	token.OR:     LOGICAL,
	token.AND:    LOGICAL,
	token.EQ:     EQUALS,
	token.NE:     EQUALS,
	token.LT:     LESSGREATER,
	token.GT:     LESSGREATER,
	token.LE:     LESSGREATER,
	token.GE:     LESSGREATER,
	token.PLUS:   SUM,
	token.MINUS:  SUM,
	token.DIV:    PRODUCT,
	token.MULT:   PRODUCT,
	token.MOD:    PRODUCT,
	token.LPAREN: CALL,
}

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.NUMBER, p.parseNumberLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.DIV, p.parseInfixExpression)
	p.registerInfix(token.MULT, p.parseInfixExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NE, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LE, p.parseInfixExpression)
	p.registerInfix(token.GE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseArrayAccess)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseRemStatement() *ast.RemStatement {
	stmt := &ast.RemStatement{Token: p.curToken}

	p.nextToken()

	comment := ""
	for !p.curTokenIs(token.EOF) && !p.curTokenIs(token.COLON) && !p.curTokenIs(token.NEWLINE) {
		comment += p.curToken.Literal + " "
		p.nextToken()
	}

	stmt.Comment = comment
	return stmt
}

func (p *Parser) parseDimStatement() *ast.DimStatement {
	stmt := &ast.DimStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Size = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return stmt
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.THEN) {
		return nil
	}

	p.nextToken()
	stmt.Consequence = p.parseStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		p.nextToken()
		stmt.Alternative = p.parseStatement()
	}

	return stmt
}

func (p *Parser) parseGotoStatement() *ast.GotoStatement {
	stmt := &ast.GotoStatement{Token: p.curToken}

	p.nextToken()
	stmt.LineNumber = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseGosubStatement() *ast.GosubStatement {
	stmt := &ast.GosubStatement{Token: p.curToken}

	p.nextToken()
	stmt.LineNumber = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	return stmt
}

func (p *Parser) parseEndStatement() *ast.EndStatement {
	stmt := &ast.EndStatement{Token: p.curToken}
	return stmt
}

func (p *Parser) parsePrintStatement() *ast.PrintStatement {
	stmt := &ast.PrintStatement{Token: p.curToken}
	stmt.Expressions = []ast.Expression{}
	stmt.Separators = []string{}
	stmt.TrailingNewline = true

	p.nextToken()

	if p.curTokenIs(token.EOF) || p.curTokenIs(token.NEWLINE) || p.curTokenIs(token.COLON) {
		return stmt
	}

	for {
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			stmt.Expressions = append(stmt.Expressions, expr)
		}

		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
			stmt.Separators = append(stmt.Separators, "")
			if p.peekTokenIs(token.EOF) || p.peekTokenIs(token.NEWLINE) || p.peekTokenIs(token.COLON) {
				stmt.TrailingNewline = false
				break
			}
			p.nextToken()
		} else if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			stmt.Separators = append(stmt.Separators, "\t")
			if p.peekTokenIs(token.EOF) || p.peekTokenIs(token.NEWLINE) || p.peekTokenIs(token.COLON) {
				stmt.TrailingNewline = false
				break
			}
			p.nextToken()
		} else {
			break
		}
	}

	return stmt
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Variable = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Start = p.parseExpression(LOWEST)

	if !p.expectPeek(token.TO) {
		return nil
	}

	p.nextToken()
	stmt.End = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.STEP) {
		p.nextToken()
		p.nextToken()
		stmt.Step = p.parseExpression(LOWEST)
	} else {
		stmt.Step = &ast.NumberLiteral{Token: token.Token{Type: token.NUMBER, Literal: "1"}, Value: 1}
	}

	return stmt
}

func (p *Parser) parseNextStatement() *ast.NextStatement {
	stmt := &ast.NextStatement{Token: p.curToken}

	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		stmt.Variable = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	return stmt
}

func (p *Parser) parseInputStatement() *ast.InputStatement {
	stmt := &ast.InputStatement{Token: p.curToken}
	stmt.Variables = []*ast.Identifier{}

	p.nextToken()

	if p.curTokenIs(token.STRING) {
		stmt.Prompt = p.curToken.Literal
		if !p.expectPeek(token.SEMICOLON) && !p.expectPeek(token.COMMA) {
			return nil
		}
		p.nextToken()
	}

	for {
		if !p.curTokenIs(token.IDENT) {
			break
		}

		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		stmt.Variables = append(stmt.Variables, ident)

		if !p.peekTokenIs(token.COMMA) {
			break
		}
		p.nextToken()
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.EOF) && !p.peekTokenIs(token.NEWLINE) && !p.peekTokenIs(token.COLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseNumberLiteral() ast.Expression {
	lit := &ast.NumberLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as number", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseArrayAccess(left ast.Expression) ast.Expression {
	arr := &ast.ArrayAccess{Token: p.curToken}

	if ident, ok := left.(*ast.Identifier); ok {
		arr.Name = ident
	} else {
		return nil
	}

	p.nextToken()
	arr.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return arr
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = make(map[int]ast.Statement)

	for !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.NEWLINE) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatementOrLine()
		if stmt != nil {
			if lineStmt, ok := stmt.(*ast.LineStatement); ok {
				program.Statements[lineStmt.LineNumber] = lineStmt.Statement
			} else {
				program.Statements[0] = stmt
			}
		}
		p.nextToken()
	}

	return program
}

// parseStatementOrLine dispatches to line or regular statement parsing.
func (p *Parser) parseStatementOrLine() ast.Statement {
	if p.curToken.Type == token.NUMBER {
		return p.parseLineStatement()
	}
	return p.parseStatement()
}

// parseStatement parses a statement and any additional statements separated by ':' on the same line.
func (p *Parser) parseStatement() ast.Statement {
	stmts := []ast.Statement{}

	stmt := p.parseSingleStatement()
	if stmt != nil {
		stmts = append(stmts, stmt)
	}

	for p.peekTokenIs(token.COLON) {
		// consume ':'
		p.nextToken()
		p.nextToken()
		if p.curTokenIs(token.NEWLINE) || p.curTokenIs(token.EOF) || p.curTokenIs(token.COLON) || p.curTokenIs(token.ELSE) {
			break
		}
		nextStmt := p.parseSingleStatement()
		if nextStmt != nil {
			stmts = append(stmts, nextStmt)
		}
	}

	if len(stmts) == 1 {
		return stmts[0]
	}

	return &ast.SequenceStatement{Statements: stmts}
}

// parseSingleStatement parses a single BASIC statement (no ':' handling).
func (p *Parser) parseSingleStatement() ast.Statement {
	switch p.curToken.Type {
	case token.PRINT:
		return p.parsePrintStatement()
	case token.LET:
		return p.parseLetStatement()
	case token.IF:
		return p.parseIfStatement()
	case token.GOTO:
		return p.parseGotoStatement()
	case token.GOSUB:
		return p.parseGosubStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.NEXT:
		return p.parseNextStatement()
	case token.INPUT:
		return p.parseInputStatement()
	case token.END:
		return p.parseEndStatement()
	case token.REM:
		return p.parseRemStatement()
	case token.DIM:
		return p.parseDimStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLineStatement() *ast.LineStatement {
	stmt := &ast.LineStatement{Token: p.curToken}

	lineNum, err := strconv.Atoi(p.curToken.Literal)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("could not parse %q as line number", p.curToken.Literal))
		return nil
	}
	stmt.LineNumber = lineNum

	p.nextToken()

	stmt.Statement = p.parseStatement()

	return stmt
}
