package token

import "fmt"

type TokenType int

const (
	EOF     TokenType = 0
	ILLEGAL           = -1

	// Identifiers and Literals
	T_IDENT        = 0x100
	T_INT          = 0x101
	T_FLOAT        = 0x102
	T_STRING       = 0x103
	T_BOOL         = 0x104
	T_QUOTED_IDENT = 0x105

	// Operators
	T_OP_EQUAL  = 0x200 // =
	T_OP_LT     = 0x201 // <
	T_OP_LT_EQ  = 0x202 // <=
	T_OP_GT     = 0x203 // >
	T_OP_GT_EQ  = 0x204 // >=
	T_OP_NOT_EQ = 0x205 // <>

	// Logical Operators
	T_KW_AND = 0x300 // and
	T_KW_OR  = 0x301 // or
	T_KW_NOT = 0x302 // not

	// build-in functions
	T_FN_HAS_PROCESS = 0x400 // has-process:

	// keywords
	T_KW_SAMPLE  = 0x700 // s:
	T_KW_PROCESS = 0x701 // p:
	T_KW_ATTR    = 0x702 // a:
	T_KW_SELECT  = 0x703 // select
	T_KW_WHERE   = 0x704 // where
	T_KW_NULL    = 0x705 // null

	// Elements
	T_LBRACKET   = 0x800 // [
	T_RBRACKET   = 0x801 // ]
	T_LPAREN     = 0x802 // (
	T_RPAREN     = 0x803 // )
	T_COMMA      = 0x804 // ,
	T_QUOTE      = 0x805 // "
	T_COLON      = 0x806 // :
	T_SEMI_COLON = 0x807 // ;
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"select":       T_KW_SELECT,
	"where":        T_KW_WHERE,
	"a:":           T_KW_ATTR,
	"p:":           T_KW_PROCESS,
	"s:":           T_KW_SAMPLE,
	"and":          T_KW_AND,
	"or":           T_KW_OR,
	"not":          T_KW_NOT,
	"null":         T_KW_NULL,
	"has-process:": T_FN_HAS_PROCESS,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	return T_IDENT
}

var tokenToStr = map[TokenType]string{
	T_OP_EQUAL:       "T_OP_EQUAL: =",
	T_OP_LT_EQ:       "T_OP_LT_EQ: <=",
	T_OP_NOT_EQ:      "T_OP_NOT_EQ: <>",
	T_OP_LT:          "T_OP_LT: <",
	T_OP_GT_EQ:       "T_OP_GT_EQ: >=",
	T_OP_GT:          "T_OP_GT: >",
	T_COMMA:          "T_COMMA: ,",
	T_LBRACKET:       "T_LBRACKET: [",
	T_RBRACKET:       "T_RBRACKET: ]",
	T_LPAREN:         "T_LPAREN: (",
	T_RPAREN:         "T_RPAREN: )",
	T_SEMI_COLON:     "T_SEMI_COLON: ;",
	T_KW_WHERE:       "T_KW_WHERE: where",
	T_KW_AND:         "T_KW_AND: and",
	T_KW_NOT:         "T_KW_NOT: not",
	T_KW_NULL:        "T_KW_NULL: null",
	T_FN_HAS_PROCESS: "T_FN_HAS_PROCESS: has-process:",
}

func TokenToStr(token TokenType) string {
	if token == T_INT {
		return "int"
	}

	if token == EOF {
		return "EOF"
	}

	if token == T_IDENT {
		return "IDENT"
	}

	if str, ok := tokenToStr[token]; ok {
		return str
	}

	return fmt.Sprintf("(%d) unknown", token)
}
