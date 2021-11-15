package parser

import (
	"fmt"
	"testing"

	"github.com/materials-commons/mql/internal/mql/ast"
	"github.com/materials-commons/mql/internal/mql/lexer"
)

func TestSimpleSelectStatement(t *testing.T) {
	input := `
	select samples;
	select processes;
	select samples, processes`

	mql := parseForTest(t, input)

	if len(mql.Statements) != 3 {
		t.Fatalf("There should have been 3 select statements, got %d", len(mql.Statements))
	}

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
	mql := parseForTest(t, input)

	if len(mql.Statements) != 3 {
		t.Fatalf("There should have been 3 select statements, got %d", len(mql.Statements))
	}

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
	input := `sa:'hardness here';`

	mql := parseForTest(t, input)

	if len(mql.Statements) != 1 {
		fmt.Printf("%+v\n", mql.Statements)
		for i, s := range mql.Statements {
			fmt.Printf("mql.Statements[%d] = %+v\n", i, s)
		}
		t.Fatalf("program has wrong number of statements. Expected 1, got = %d", len(mql.Statements))
	}

	es, ok := mql.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Expected *ast.ExpressionStatement, got %T", mql.Statements[0])
	}

	sampleAttribute, ok := es.Expression.(*ast.SampleAttributeIdentifier)
	if !ok {
		t.Fatalf("Expected *ast.SampleAttributeIdentifier, got %T", es.Expression)
	}

	if sampleAttribute.Value != "hardness here" {
		t.Errorf("Expected value 'hardness here', got %s", sampleAttribute.Value)
	}

	if sampleAttribute.TokenLiteral() != "hardness here" {
		t.Errorf("Expected token literal 'hardness here', got %s", sampleAttribute.Value)

	}
}

func parseForTest(t *testing.T, input string) *ast.MQL {
	l := lexer.New(input)
	p := New(l)

	mql := p.ParseMQL()
	if mql == nil {
		t.Fatalf("ParseMQL returned nil")
	}

	return mql
}
