package ast

import "github.com/materials-commons/mql/internal/mql/token"

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type MQL struct {
	Statements []Statement
}

type SelectStatement struct {
	Token               token.Token
	SelectionStatements []Statement
	WhereStatement      WhereStatement
}

func (s *SelectStatement) statementNode() {
}

func (s *SelectStatement) TokenLiteral() string {
	return s.Token.Literal
}

func (s *SelectStatement) String() string {
	return ""
}

type WhereStatement struct {
	Token    token.Token
	Criteria []Statement
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (l *IntegerLiteral) expressionNode() {
}

func (l *IntegerLiteral) TokenLiteral() string {
	return l.Token.Literal
}
func (l *IntegerLiteral) String() string {
	return l.Token.Literal
}

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (l *FloatLiteral) expressionNode() {
}

func (l *FloatLiteral) TokenLiteral() string {
	return l.Token.Literal
}
func (l *FloatLiteral) String() string {
	return l.Token.Literal
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (l *StringLiteral) expressionNode() {
}

func (l *StringLiteral) TokenLiteral() string {
	return l.Token.Literal
}

func (l *StringLiteral) String() string {
	return l.Token.Literal
}
