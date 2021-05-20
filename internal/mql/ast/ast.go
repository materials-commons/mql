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

func (s *SelectStatement) statementNode()       {}
func (s *SelectStatement) TokenLiteral() string { return s.Token.Literal }
func (s *SelectStatement) String() string {
	return ""
}

type WhereStatement struct {
	Token    token.Token
	Criteria []Statement
}
