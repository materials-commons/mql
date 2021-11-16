package parser

import (
	"testing"

	"github.com/materials-commons/mql/internal/mql/ast"
	"github.com/materials-commons/mql/internal/mql/lexer"
)

func TestSimpleSelectStatement(t *testing.T) {
	input := `
	select samples;
	select processes;
	select samples, processes`

	mql := parseForTest(t, input, 3)

	tests := []struct {
		isSamplesSelection   bool
		isProcessesSelection bool
	}{
		{isSamplesSelection: true, isProcessesSelection: false},
		{isSamplesSelection: false, isProcessesSelection: true},
		{isSamplesSelection: true, isProcessesSelection: true},
	}
	for i, tt := range tests {
		stmt := mql.Statements[i]
		if !testSimpleSelectStatement(t, stmt, tt.isSamplesSelection, tt.isProcessesSelection) {
			return
		}
	}
}

func testSimpleSelectStatement(t *testing.T, s ast.Statement, isSamplesSelection, isProcessesSelection bool) bool {
	if s.TokenLiteral() != "select" {
		t.Errorf("s.TokenLiteral not 'select', got=%q", s.TokenLiteral())
		return false
	}
	selectStatement, ok := s.(*ast.SelectStatement)
	if !ok {
		t.Errorf("statement not *ast.SelectStatement, got=%T", selectStatement)
		return false
	}
	for _, selection := range selectStatement.SelectionStatements {
		switch st := selection.(type) {
		case *ast.SamplesSelectionStatement:
			if !isSamplesSelection {
				t.Errorf("Found samples selection, but was not expecting it")
				return false
			}
		case *ast.ProcessesSelectionStatement:
			if !isProcessesSelection {
				t.Errorf("Found processes selection, but was not expecting it")
				return false
			}
		default:
			t.Errorf("Expected *ast.{SamplesSelectionStatement,ProcessesSelectionStatement, got %T", st)
			return false
		}
	}
	return true
}

func TestSelectWithWhereStatement(t *testing.T) {
	input := `
	select samples where;
	select processes where;
	select samples, processes where;
`
	mql := parseForTest(t, input, 3)

	for _, s := range mql.Statements {
		if s.TokenLiteral() != "select" {
			t.Fatalf("s.TokenLiteral not 'select', got=%q", s.TokenLiteral())
		}
		selectStatement, ok := s.(*ast.SelectStatement)
		if !ok {
			t.Fatalf("statement not *ast.SelectStatement, got=%T", selectStatement)
		}
		if selectStatement.WhereStatement == nil {
			t.Fatalf("Statement should have had a where statement, but didn't")
		}
	}
}

func TestSampleIdentifierExpression(t *testing.T) {
	input := `sa:hardness;
		sample-attr:hardness;
		sample:hardness;
		s:hardness;
		sample:'with space';`

	tests := []struct {
		token      string
		identifier string
	}{
		{"sa:", "hardness"},
		{"sample-attr:", "hardness"},
		{"sample:", "hardness"},
		{"s:", "hardness"},
		{"sample:", "with space"},
	}

	mql := parseForTest(t, input, len(tests))

	for i, test := range tests {
		testSampleIdentifierExpression(t, mql.Statements[i], test.token, test.identifier)
	}

}

func testSampleIdentifierExpression(t *testing.T, s ast.Statement, token, identifier string) {
	es, ok := s.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected *ast.ExpressionStatement, got %T", s)
	}

	if es.TokenLiteral() != token {
		t.Fatalf("Expect token %s, got %s", token, es.TokenLiteral())
	}

	sampleAttribute, ok := es.Expression.(*ast.SampleAttributeIdentifier)
	if !ok {
		t.Fatalf("Expected *ast.SampleAttributeIdentifier, got %T", es.Expression)
	}

	if sampleAttribute.Value != identifier {
		t.Errorf("Expected value '%s', got %s", identifier, sampleAttribute.Value)
	}

	if sampleAttribute.TokenLiteral() != identifier {
		t.Errorf("Expected token literal '%s', got %s", identifier, sampleAttribute.Value)
	}
}

func TestProcessIdentifierExpression(t *testing.T) {
	input := `pa:hardness;
		process-attr:hardness;
		process:hardness;
		p:hardness;
		process:'with space';`

	tests := []struct {
		token      string
		identifier string
	}{
		{"pa:", "hardness"},
		{"process-attr:", "hardness"},
		{"process:", "hardness"},
		{"p:", "hardness"},
		{"process:", "with space"},
	}

	mql := parseForTest(t, input, len(tests))

	for i, test := range tests {
		testProcessIdentifierExpression(t, mql.Statements[i], test.token, test.identifier)
	}

}

func testProcessIdentifierExpression(t *testing.T, s ast.Statement, token, identifier string) {
	es, ok := s.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected *ast.ExpressionStatement, got %T", s)
	}

	if es.TokenLiteral() != token {
		t.Fatalf("Expect token %s, got %s", token, es.TokenLiteral())
	}

	processAttribute, ok := es.Expression.(*ast.ProcessAttributeIdentifier)
	if !ok {
		t.Fatalf("Expected *ast.ProcessAttributeIdentifier, got %T", es.Expression)
	}

	if processAttribute.Value != identifier {
		t.Errorf("Expected value '%s', got %s", identifier, processAttribute.Value)
	}

	if processAttribute.TokenLiteral() != identifier {
		t.Errorf("Expected token literal '%s', got %s", identifier, processAttribute.Value)
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"
	mql := parseForTest(t, input, 1)

	statement, ok := mql.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement not *ast.ExpressionStatement, got=%T", statement)
	}

	literal, ok := statement.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("literal not *ast.IntegerLiteral, got %T", literal)
	}

	if literal.Value != 5 {
		t.Fatalf("expected 5 for value got %d", literal.Value)
	}

	if literal.TokenLiteral() != "5" {
		t.Fatalf("expected token '5' got '%s'", literal.TokenLiteral())
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"-5", "-", 5},
	}

	for _, test := range tests {
		mql := parseForTest(t, test.input, 1)
		expression := checkForExpressionStatement(t, mql.Statements[0])
		prefixExpression := checkForPrefixExpression(t, expression.Expression)
		if prefixExpression.Operator != test.operator {
			t.Fatalf("expected prefixExpression.Operator '%s', got '%s'", test.operator, prefixExpression.Operator)
		}
	}
}

func checkForExpressionStatement(t *testing.T, statement ast.Statement) *ast.ExpressionStatement {
	s, ok := statement.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("statement not *ast.ExpressionStatement, got=%T", s)
	}

	return s
}

func checkForPrefixExpression(t *testing.T, expression ast.Expression) *ast.PrefixExpression {
	e, ok := expression.(*ast.PrefixExpression)
	if !ok {
		t.Fatalf("statement not *ast.ExpressionStatement, got=%T", e)
	}

	return e
}

func parseForTest(t *testing.T, input string, length int) *ast.MQL {
	l := lexer.New(input)
	p := New(l)

	mql := p.ParseMQL()
	if mql == nil {
		t.Fatalf("ParseMQL returned nil")
	}

	if len(mql.Statements) != length {
		t.Fatalf("program has wrong number of statements. Expected %d, got = %d", length, len(mql.Statements))
	}
	return mql
}
