package lexer

import (
	"testing"

	"github.com/materials-commons/mql/internal/mql/token"
)

func TestNextToken(t *testing.T) {
	input := `select p:[name, a:time], s:[a:'metal hardness'] where and not null`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.T_KW_SELECT, "select"},
		{token.T_KW_PROCESS, "p:"},
		{token.T_LBRACKET, "["},
		{token.T_IDENT, "name"},
		{token.T_COMMA, ","},
		{token.T_KW_ATTR, "a:"},
		{token.T_IDENT, "time"},
		{token.T_RBRACKET, "]"},
		{token.T_COMMA, ","},
		{token.T_KW_SAMPLE, "s:"},
		{token.T_LBRACKET, "["},
		{token.T_KW_ATTR, "a:"},
		{token.T_IDENT, "metal hardness"},
		{token.T_RBRACKET, "]"},
		{token.T_KW_WHERE, "where"},
		{token.T_KW_AND, "and"},
		{token.T_KW_NOT, "not"},
		{token.T_KW_NULL, "null"},
	}

	l := New(input)
	for i, test := range tests {
		tok := l.NextToken()
		if tok.Type != test.expectedType {
			t.Fatalf("tests[%d] - Token Type wrong. Expected='%s', got='%s': %s", i,
				token.TokenToStr(test.expectedType), token.TokenToStr(tok.Type), tok.Literal)
		}

		if tok.Literal != test.expectedLiteral {
			t.Fatalf("tests[%d] - Literal wrong. Expected=%q, got=%q", i, test.expectedLiteral, tok.Literal)
		}
	}
}
