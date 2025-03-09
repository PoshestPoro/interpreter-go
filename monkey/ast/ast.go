package ast

import "monkey/token"

type node interface {
	TokenLiteral() string
}

type Statement interface {
	node
	statement_node()
}

type expression interface {
	node
	expression_node()
}

type Program struct {
	Statements []Statement
}

type Let_statement struct {
	Token token.Token
	Name  *Identifier
	Value expression
}

type Return_statement struct {
	Token token.Token
	Value expression
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
