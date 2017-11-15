package ast

import (
	"bytes"
	"fmt"

	"github.com/abates/monkey-go/token"
)

func writeString(str ...interface{}) string {
	var buffer bytes.Buffer
	for _, s := range str {
		buffer.WriteString(fmt.Sprintf("%v", s))
	}

	return buffer.String()
}

type statement struct{}

func (statement) statementNode() {}

type expression struct{}

func (expression) expressionNode() {}

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

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	stmts := make([]interface{}, len(p.Statements))
	for i, stmt := range p.Statements {
		stmts[i] = stmt
	}
	return writeString(stmts...)
}

type LetStatement struct {
	token.Token
	statement
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) String() string {
	if ls.Value != nil {
		return writeString(ls.TokenLiteral(), " ", ls.Name, " = ", ls.Value, ";")
	}
	return writeString(ls.TokenLiteral(), " = ", ";")
}

type Identifier struct {
	token.Token
	expression
	Value string
}

func (i *Identifier) String() string { return i.Value }

type ReturnStatement struct {
	token.Token
	statement
	ReturnValue Expression
}

func (rs *ReturnStatement) String() string {
	if rs.ReturnValue != nil {
		return writeString(rs.TokenLiteral(), " ", rs.ReturnValue, ";")
	}
	return writeString(rs.TokenLiteral(), ";")
}

type ExpressionStatement struct {
	token.Token
	statement
	Expression Expression
}

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type IntegerLiteral struct {
	token.Token
	expression
	Value int64
}

func (il *IntegerLiteral) String() string { return il.Token.Literal }

type PrefixExpression struct {
	token.Token
	expression
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) String() string {
	return writeString("(", pe.Operator, pe.Right, ")")
}

type InfixExpression struct {
	token.Token
	expression
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) String() string {
	return writeString("(", ie.Left, " ", ie.Operator, " ", ie.Right, ")")
}
