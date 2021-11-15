package parser

import (
	"fmt"
	"strconv"

	"github.com/materials-commons/mql/internal/mql"
	"github.com/materials-commons/mql/internal/mql/ast"
	"github.com/materials-commons/mql/internal/mql/lexer"
	"github.com/materials-commons/mql/internal/mql/token"
)

// Precendence from lowest to highest
const (
	_ int = iota
	LOWEST
	EQUALS      // =
	LESSGREATER // > or < or <= or >=
	BOOLEAN
)

var precendences = map[token.TokenType]int{
	token.EQUAL: EQUALS,
	token.NOTEQ: EQUALS,
	token.LT:    LESSGREATER,
	token.LTEQ:  LESSGREATER,
	token.GT:    LESSGREATER,
	token.GTEQ:  LESSGREATER,
	token.AND:   BOOLEAN,
	token.NOT:   BOOLEAN,
	token.OR:    BOOLEAN,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l              *lexer.Lexer
	errors         []string
	curToken       token.Token
	peekToken      token.Token
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.infixParseFns = make(map[token.TokenType]infixParseFn)

	// Prefix
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.SAMPLES, p.parseSamplesLiteral)
	p.registerPrefix(token.PROCESSES, p.parseProcessesLiteral)
	p.registerPrefix(token.SAMPLE_HAS_ATTRIBUTE_FUNC, p.parseSampleHasAttributeFunc)
	p.registerPrefix(token.SAMPLE_HAS_PROCESS_FUNC, p.parseSampleHasProcessFunc)
	p.registerPrefix(token.PROCESS_HAS_ATTRIBUTE_FUNC, p.parseProcessHasAttributeFunc)
	p.registerPrefix(token.PROCESS_HAS_SAMPLE_FUNC, p.parseProcessHasSampleFunc)
	p.registerPrefix(token.SAMPLE_ATTR, p.parseSampleAttrFunc)

	// Infix
	p.registerInfix(token.AND, p.parseAndExpression)
	p.registerInfix(token.OR, p.parseOrExpression)

	// Read two tokens so that currentToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) ParseMQL() *ast.MQL {
	mqlProgram := &ast.MQL{}
	mqlProgram.Statements = []ast.Statement{}
	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			fmt.Println("  adding statement")
			mqlProgram.Statements = append(mqlProgram.Statements, stmt)
		}
		p.nextToken()
	}

	return mqlProgram
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.SELECT:
		return p.parseSelectStatement()
	//case token.SEMICOLON:
	//	return
	default:
		// Allow expression parsing, at least for now
		return p.parseExpressionStatement()
		// error here for now
		//log.Fatalf("Top level statement can only be a select")
		//return nil
	}
}

func (p *Parser) parseSelectStatement() ast.Statement {
	statement := &ast.SelectStatement{Token: p.curToken, SelectionStatements: []ast.Statement{}}
	if !p.peekTokenIs(token.SAMPLES) && !p.peekTokenIs(token.PROCESSES) {
		return nil
	}
	p.nextToken()
	statement.SelectionStatements = p.parseSelectionStatement()
	if p.curTokenIs(token.WHERE) {
		statement.WhereStatement = p.parseWhereStatement()
	}

	return statement
}

func (p *Parser) parseSelectionStatement() []ast.Statement {
	var selectionStatements []ast.Statement
	for {
		switch {
		case p.curTokenIs(token.SAMPLES):
			selectionStatements = append(selectionStatements, &ast.SamplesSelectionStatement{Token: p.curToken})
		case p.curTokenIs(token.PROCESSES):
			selectionStatements = append(selectionStatements, &ast.ProcessesSelectionStatement{Token: p.curToken})
		case p.curTokenIs(token.COMMA):
			// skip over to next token
		default:
			return selectionStatements
		}
		p.nextToken()
	}
}

func (p *Parser) parseWhereStatement() *ast.WhereStatement {
	whereStatement := &ast.WhereStatement{Statements: []ast.Statement{}}

	// move past where
	p.nextToken()

	for !p.curTokenIs(token.SEMICOLON) {
		// TODO: Skip parsing expressions until we encounter a semicolon
		p.nextToken()
	}

	return whereStatement
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	fmt.Println("parseExpressionStatement")
	statement := &ast.ExpressionStatement{Token: p.curToken}
	statement.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return statement
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	fmt.Printf("parseExpression p.curToken.Type = %d/%s\n", p.curToken.Type, p.curToken.Literal)
	prefixFn := p.prefixParseFns[p.curToken.Type]
	if prefixFn == nil {
		fmt.Printf("No function found for %d/%s\n", p.curToken.Type, p.curToken.Literal)
		p.appendError("no prefix parse function for %s found", token.TokenToStr(p.curToken.Type))
		return nil
	}

	leftExp := prefixFn()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infixFn := p.infixParseFns[p.peekToken.Type]
		if infixFn == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infixFn(leftExp)
	}

	return leftExp
}

func (p *Parser) parseSampleAttrFunc() ast.Expression {
	fmt.Println("parseSampleAttrFunc p.peekToken = ", p.peekToken.Literal)
	p.nextToken()

	return &ast.SampleAttributeIdentifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) appendError(msg string, args ...interface{}) {
	p.errors = append(p.errors, fmt.Sprintf(msg, args...))
}

func (p *Parser) registerPrefix(t token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[t] = fn
}

func (p *Parser) registerInfix(t token.TokenType, fn infixParseFn) {
	p.infixParseFns[t] = fn
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	var err error
	literal := &ast.IntegerLiteral{Token: p.curToken}
	if literal.Value, err = strconv.ParseInt(p.curToken.Literal, 0, 64); err != nil {
		p.appendError("could not parse %q as integer", p.curToken.Literal)
		return nil
	}

	return literal
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	var err error
	literal := &ast.FloatLiteral{Token: p.curToken}
	if literal.Value, err = strconv.ParseFloat(p.curToken.Literal, 64); err != nil {
		p.appendError("could not parse %q as float", p.curToken.Literal)
		return nil
	}

	return literal
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

//func (p *Parser) ParseMQL() *ast.MQL {
//	mql := &ast.MQL{}
//	mql.Statements = []ast.Statement{}
//	for !p.curTokenIs(token.EOF) {
//		statement := p.parseStatement()
//		if statement != nil {
//			mql.Statements = append(mql.Statements, statement)
//		}
//		p.nextToken()
//	}
//	return mql
//}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
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

func (p *Parser) peekPrecedence() int {
	return p.getPrecedence(p.peekToken.Type)
}

func (p *Parser) curPrecedence() int {
	return p.getPrecedence(p.curToken.Type)
}

func (p *Parser) getPrecedence(t token.TokenType) int {
	if p, ok := precendences[t]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) peekError(t token.TokenType) {
	p.errors = append(p.errors, fmt.Sprintf("Expect next token to be %s, got %s instead", token.TokenToStr(t),
		token.TokenToStr(p.peekToken.Type)))
}

func (p *Parser) parseAndExpression(expression ast.Expression) ast.Expression {
	s := mql.AndStatement{
		Left:  nil,
		Right: nil,
	}
	_ = s
	return nil
}

func (p *Parser) parseSamplesLiteral() ast.Expression {
	return nil
}

func (p *Parser) parseProcessesLiteral() ast.Expression {
	return nil
}

func (p *Parser) parseSampleHasAttributeFunc() ast.Expression {
	return nil
}

func (p *Parser) parseSampleHasProcessFunc() ast.Expression {
	return nil
}

func (p *Parser) parseProcessHasAttributeFunc() ast.Expression {
	return nil
}

func (p *Parser) parseProcessHasSampleFunc() ast.Expression {
	return nil
}

func (p *Parser) parseOrExpression(expression ast.Expression) ast.Expression {
	return nil
}
