package evaluator

import (
	"github.com/materials-commons/mql/internal/mql/ast"
	"github.com/materials-commons/mql/internal/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.MQL:
		return evalMQL(node)
	case *ast.SelectStatement:
		return evalSelectStatement(node)
	case *ast.WhereStatement:
		return evalWhereStatement(node)
	}

	return nil
}

func evalMQL(mql *ast.MQL) object.Object {
	var result object.Object

	for _, statement := range mql.Statements {
		result = Eval(statement)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalSelectStatement(s *ast.SelectStatement) object.Object {
	return nil
}

func evalWhereStatement(s *ast.WhereStatement) object.Object {
	return nil
}
