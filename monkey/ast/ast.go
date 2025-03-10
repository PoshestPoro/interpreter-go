package ast

import (
	"bytes"
	"monkey/token"
)

type Statement interface {
	Node
	statement_node()
}

type Expression interface {
	Node
	expression_node()
}

type Program struct {
	Statements []Statement
}

type Let_statement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

type Return_statement struct {
	Token        token.Token
	Return_value Expression
}

type Integer_literal struct {
	token.Token
	Value int64
}

type Prefix_expression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

type Infix_expression struct {
	Token    token.Token
	Operator string
	Right    Expression
	Left     Expression
}

type Expression_statement struct {
	Token      token.Token
	Expression Expression
}

type Node interface {
	TokenLiteral() string
	String() string
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (ls *Let_statement) statement_node()      {}
func (ls *Let_statement) TokenLiteral() string { return ls.Token.Literal }

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expression_node()     {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

func (rs *Return_statement) statement_node()      {}
func (rs *Return_statement) TokenLiteral() string { return rs.Token.Literal }

func (ex *Expression_statement) statement_node()      {}
func (ex *Expression_statement) TokenLiteral() string { return ex.Token.Literal }

func (il *Integer_literal) expression_node()     {}
func (il *Integer_literal) TokenLiteral() string { return il.Token.Literal }

func (pe *Prefix_expression) expression_node()     {}
func (pe *Prefix_expression) TokenLiteral() string { return pe.Token.Literal }

func (ie *Infix_expression) expression_node()     {}
func (ie *Infix_expression) TokenLiteral() string { return ie.Token.Literal }

func (ie Infix_expression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

func (il *Integer_literal) String() string { return il.Token.Literal }

func (pe *Prefix_expression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())

	}
	return out.String()
}

func (ls *Let_statement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")
	return out.String()
}

func (rs *Return_statement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")
	if rs.Return_value != nil {
		out.WriteString(rs.Return_value.String())
	}

	out.WriteString(";")

	return out.String()
}

func (es *Expression_statement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

func (i *Identifier) String() string { return i.Value }
