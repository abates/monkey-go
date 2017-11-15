package parser

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/abates/monkey-go/ast"
	"github.com/abates/monkey-go/lexer"
)

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}

	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("s.Name not '%s'. got=%s", name, letStmt.Name)
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

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5;", 5},
		{"return true;", true},
		{"return foobar;", "foobar"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d", len(program.Statements))
		}

		stmt := program.Statements[0]
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Fatalf("stmt not *ast.returnStatement. got=%T", stmt)
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Fatalf("returnStmt.TokenLiteral not 'return', got %q", returnStmt.TokenLiteral())
		}
	}
}

func testExpression(t *testing.T, input string, typ reflect.Type) ast.Expression {
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("%q program has the wrong number of statements. got=%d", input, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("%q program.Statements[0] is not ast.ExpressionStatement. got=%T", input, program.Statements[0])
		return nil
	}

	if reflect.TypeOf(stmt.Expression) != typ {
		t.Fatalf("expression not %s. got=%s", typ, reflect.TypeOf(stmt.Expression))
	}

	return stmt.Expression
}

func testPrimitive(test interface{}) func(*testing.T, interface{}) {
	return func(t *testing.T, value interface{}) {
		if test != value {
			t.Errorf("Expected %v(%T) but got %v(%T)", test, test, value, value)
		}
	}
}

func testIntLiteral(test int64) func(*testing.T, interface{}) {
	return func(t *testing.T, value interface{}) {
		if literal, ok := value.(*ast.IntegerLiteral); ok {
			if literal.Value != test {
				t.Errorf("Expected IntegerLiteral value to be %d but got %d", test, literal.Value)
			}

			if literal.TokenLiteral() != fmt.Sprintf("%d", test) {
				t.Errorf("Expected IntegerLiteral token to be \"%d\" but got %q", test, literal.TokenLiteral())
			}
		} else {
			t.Errorf("Expected *ast.IntegerLiteral but got %T", value)
		}
	}
}

func TestExpressions(t *testing.T) {
	tests := []struct {
		input      string
		token      string
		typ        reflect.Type
		valueField string
		compare    func(*testing.T, interface{})
	}{
		{"foobar", "foobar", reflect.TypeOf(&ast.Identifier{}), "Value", testPrimitive("foobar")},
		{"5;", "5", reflect.TypeOf(&ast.IntegerLiteral{}), "Value", testPrimitive(int64(5))},
		{"!5;", "!", reflect.TypeOf(&ast.PrefixExpression{}), "Right", testIntLiteral(5)},
		{"-15;", "-", reflect.TypeOf(&ast.PrefixExpression{}), "Right", testIntLiteral(15)},
	}

	for _, test := range tests {
		expression := testExpression(t, test.input, test.typ)
		if expression.TokenLiteral() != test.token {
			t.Errorf("expression.TokenLiteral not %s. got=%s", test.token, expression.TokenLiteral())
		}

		value := reflect.Indirect(reflect.ValueOf(expression))
		valueField := value.FieldByName(test.valueField)
		test.compare(t, valueField.Interface())
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	tests := []struct {
		input      string
		leftValue  int64
		operator   string
		rightValue int64
	}{
		{"5 + 5", 5, "+", 5},
		{"5 - 5", 5, "-", 5},
		{"5 * 5", 5, "*", 5},
		{"5 / 5", 5, "/", 5},
		{"5 > 5", 5, ">", 5},
		{"5 < 5", 5, "<", 5},
		{"5 == 5", 5, "==", 5},
		{"5 != 5", 5, "!=", 5},
	}

	for _, test := range tests {
		expression, _ := testExpression(t, test.input, reflect.TypeOf(&ast.InfixExpression{})).(*ast.InfixExpression)
		if expression.Operator != test.operator {
			t.Errorf("Expected operator %s but got %s", test.operator, expression.Operator)
		}
		testIntLiteral(test.leftValue)(t, expression.Left)
		testIntLiteral(test.rightValue)(t, expression.Right)
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"-a * b", "((-a) * b)"},
		{"!-a", "(!(-a))"},
		{"a + b + c", "((a + b) + c)"},
		{"a + b - c", "((a + b) - c)"},
		{"a * b * c", "((a * b) * c)"},
		{"a * b / c", "((a * b) / c)"},
		{"a + b * c + d / e - f", "(((a + (b * c)) + (d / e)) - f)"},
		{"3 + 4; -5 * 5", "(3 + 4)((-5) * 5)"},
		{"5 > 4 == 3 < 4", "((5 > 4) == (3 < 4))"},
		{"5 < 4 != 3 > 4", "((5 < 4) != (3 > 4))"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5", "((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
	}

	for i, test := range tests {
		l := lexer.New(test.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if test.expected != actual {
			t.Errorf("test[%d] expected=%q, got=%q", i, test.expected, actual)
		}
	}
}
