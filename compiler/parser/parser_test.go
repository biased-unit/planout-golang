package parser

import (
	"testing"

	"github.com/stitchfix/planout-golang/compiler/ast"
	"github.com/stitchfix/planout-golang/compiler/lexer"
)

func TestAssignmentStatement(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{
			name:               "assign integer literal",
			input:              "x = 5;",
			expectedIdentifier: "x",
			expectedValue:      5,
		},
		{
			name:               "assign float literal",
			input:              "x = 3.14;",
			expectedIdentifier: "x",
			expectedValue:      3.14,
		},
		{
			name:               "assign exponential literal",
			input:              "my_exp = 3.14E-9;",
			expectedIdentifier: "my_exp",
			expectedValue:      3.14E-9,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.New(tc.input)
			p := New(l)

			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Seq) != 1 {
				t.Fatalf("program does not contain 1 statements. got=%d}", len(program.Seq))
			}

			stmt := program.Seq[0]
			if !testAssignStatement(t, stmt, tc.expectedIdentifier) {
				return
			}

			val := stmt.(*ast.AssignmentStatement).Value
			if !testLiteralExpression(t, val, tc.expectedValue) {
				return
			}

		})

	}
}

func TestReturnStatements(t *testing.T) {
	input := `
		return 5;
		return 3.14;
		return 3.14E+09;`
	expectedValues := []interface{}{5, 3.14, 3.14E9}

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Seq) != 3 {
		t.Fatalf("program does not contain 3 statements. got=%d)}", len(program.Seq))
	}

	for i, exp := range program.Seq {
		returnExp, ok := exp.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("exp not *ast.ReturnExpresion. got=%T", exp)
		}
		testLiteralExpression(t, returnExp.Value, expectedValues[i])
	}
}

func TestIdentiferExpression(t *testing.T) {
	input := "x=my_var;"
	s := lexer.New(input)
	p := New(s)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Seq) != 1 {
		t.Fatalf("program does not have 1 expression. got=%d", len(program.Seq))
	}
	exp, ok := program.Seq[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("program.Seq[0] is not ast.AssignmentStatement. got=%T",
			program.Seq[0])
	}

	ident, ok := exp.Value.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp.Value not *ast.Identifier. got=%T", exp.Value)
	}
	if ident.Var != "my_var" {
		t.Fatalf("ident.Var not %s. got=%s", "my_var", ident.Var)
	}
}

func testAssignStatement(t *testing.T, s ast.Statement, name string) bool {
	stmt, ok := s.(*ast.AssignmentStatement)
	if !ok {
		t.Errorf("lx not *ast.AssignStatement. got=%T", s)
		return false
	}

	if stmt.Var != name {
		t.Errorf("stmt.Var not '%s'. got=%s", name, stmt.Var)
		return false
	}

	return true
}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case float64:
		return testFloatLiteral(t, exp, v)
		//case string:
		//	return testIdentifier(t, exp, v)
		//case bool:
		//	return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testIntegerLiteral(t *testing.T, exp ast.Expression, value int64) bool {
	il, ok := exp.(ast.IntegerLiteral)
	if !ok {
		t.Errorf("exp not ast.IntegerLiteral. got=%T", exp)
		return false
	}

	if int64(il) != value {
		t.Errorf("il.Value not %d. got=%d", value, int64(il))
		return false
	}

	return true
}

func testFloatLiteral(t *testing.T, exp ast.Expression, value float64) bool {
	fl, ok := exp.(ast.FloatLiteral)
	if !ok {
		t.Errorf("il not ast.FloatLiteral. got=%T", exp)
		return false
	}

	if float64(fl) != value {
		t.Errorf("fl not %g. got=%g", value, float64(fl))
		return false
	}

	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
