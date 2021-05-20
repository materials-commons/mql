package parser

import (
	"fmt"
	"log"

	"github.com/materials-commons/mql/internal/mql/ast"
	"github.com/materials-commons/mql/internal/mql/lexer"
	"github.com/materials-commons/mql/internal/mql/token"
)

type Parser struct {
	l         *lexer.Lexer
	errors    []string
	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	// Read two tokens so that currentToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseMQL() *ast.MQL {
	mql := &ast.MQL{}
	mql.Statements = []ast.Statement{}
	for !p.curTokenIs(token.EOF) {
		statement := p.parseStatement()
		if statement != nil {
			mql.Statements = append(mql.Statements, statement)
		}
		p.nextToken()
	}

	return mql
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.T_KW_SELECT:
		return p.parseSelectStatement()
	default:
		// error here for now
		log.Fatalf("Top level statement can only be a select")
		return nil
	}
}

func (p *Parser) parseSelectStatement() ast.Statement {
	statement := &ast.SelectStatement{Token: p.curToken, SelectionStatements: []ast.Statement{}}
	p.nextToken()
	statement.SelectionStatements = p.parseSelectionStatements()
	if !p.expectPeek(token.T_KW_WHERE) {
		return statement
	}
	return nil
}

func (p *Parser) parseSelectionStatements() []ast.Statement {
	return nil
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) peekError(t token.TokenType) {
	p.errors = append(p.errors, fmt.Sprintf("Expect next token to be %s, got %s instead", token.TokenToStr(t),
		token.TokenToStr(p.peekToken.Type)))
}
