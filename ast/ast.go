package ast

import "github.com/basis-ex/token"

type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements map[int]Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		for _, stmt := range p.Statements {
			return stmt.TokenLiteral()
		}
	}
	return ""
}

type LineStatement struct {
	Token      token.Token
	LineNumber int
	Statement  Statement
}

func (ls *LineStatement) statementNode()       {}
func (ls *LineStatement) TokenLiteral() string { return ls.Token.Literal }

// SequenceStatement represents multiple statements on a single BASIC line separated by ':'.
type SequenceStatement struct {
	Statements []Statement
}

func (ss *SequenceStatement) statementNode()       {}
func (ss *SequenceStatement) TokenLiteral() string { return "" }

type PrintStatement struct {
	Token           token.Token
	Expressions     []Expression
	Separators      []string
	TrailingNewline bool
}

func (ps *PrintStatement) statementNode()       {}
func (ps *PrintStatement) TokenLiteral() string { return ps.Token.Literal }

type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

type IfStatement struct {
	Token       token.Token
	Condition   Expression
	Consequence Statement
	Alternative Statement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }

type GotoStatement struct {
	Token      token.Token
	LineNumber Expression
}

func (gs *GotoStatement) statementNode()       {}
func (gs *GotoStatement) TokenLiteral() string { return gs.Token.Literal }

type GosubStatement struct {
	Token      token.Token
	LineNumber Expression
}

func (gs *GosubStatement) statementNode()       {}
func (gs *GosubStatement) TokenLiteral() string { return gs.Token.Literal }

type ReturnStatement struct {
	Token token.Token
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

type ForStatement struct {
	Token     token.Token
	Variable  *Identifier
	Start     Expression
	End       Expression
	Step      Expression
	Body      []Statement
	LineStart int
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }

type NextStatement struct {
	Token    token.Token
	Variable *Identifier
}

func (ns *NextStatement) statementNode()       {}
func (ns *NextStatement) TokenLiteral() string { return ns.Token.Literal }

type InputStatement struct {
	Token     token.Token
	Prompt    string
	Variables []*Identifier
}

func (is *InputStatement) statementNode()       {}
func (is *InputStatement) TokenLiteral() string { return is.Token.Literal }

type EndStatement struct {
	Token token.Token
}

func (es *EndStatement) statementNode()       {}
func (es *EndStatement) TokenLiteral() string { return es.Token.Literal }

type RemStatement struct {
	Token   token.Token
	Comment string
}

func (rs *RemStatement) statementNode()       {}
func (rs *RemStatement) TokenLiteral() string { return rs.Token.Literal }

type DimStatement struct {
	Token token.Token
	Name  *Identifier
	Size  Expression
}

func (ds *DimStatement) statementNode()       {}
func (ds *DimStatement) TokenLiteral() string { return ds.Token.Literal }

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

type NumberLiteral struct {
	Token token.Token
	Value float64
}

func (nl *NumberLiteral) expressionNode()      {}
func (nl *NumberLiteral) TokenLiteral() string { return nl.Token.Literal }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

type ArrayAccess struct {
	Token token.Token
	Name  *Identifier
	Index Expression
}

func (aa *ArrayAccess) expressionNode()      {}
func (aa *ArrayAccess) TokenLiteral() string { return aa.Token.Literal }
