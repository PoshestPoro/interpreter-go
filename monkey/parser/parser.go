package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

type (
	prefix_Parse_Fn func() ast.Expression
	infix_Parse_fn  func(ast.Expression) ast.Expression
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
}

type Parser struct {
	l *lexer.Lexer

	cur_token  token.Token
	peek_token token.Token

	errors []string

	prefix_Parse_Fns map[token.TokenType]prefix_Parse_Fn
	infix_Parse_Fns  map[token.TokenType]infix_Parse_fn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	p.next_token()
	p.next_token()

	p.prefix_Parse_Fns = make(map[token.TokenType]prefix_Parse_Fn)
	p.infix_Parse_Fns = make(map[token.TokenType]infix_Parse_fn)

	p.register_prefix(token.IDENT, p.parse_identifier)
	p.register_prefix(token.INT, p.parse_integer_literal)
	p.register_prefix(token.BANG, p.parse_prefix_expression)
	p.register_prefix(token.MINUS, p.parse_prefix_expression)

	p.register_infix(token.PLUS, p.parse_infix_expression)
	p.register_infix(token.MINUS, p.parse_infix_expression)
	p.register_infix(token.SLASH, p.parse_infix_expression)
	p.register_infix(token.ASTERISK, p.parse_infix_expression)
	p.register_infix(token.EQ, p.parse_infix_expression)
	p.register_infix(token.NOT_EQ, p.parse_infix_expression)
	p.register_infix(token.LT, p.parse_infix_expression)
	p.register_infix(token.GT, p.parse_infix_expression)

	return p
}

func (p *Parser) parse_infix_expression(left ast.Expression) ast.Expression {
	defer untrace(trace("prefix_infix_expression"))

	expr := &ast.Infix_expression{
		Token:    p.cur_token,
		Operator: p.cur_token.Literal,
		Left:     left}
	precedence := p.cur_precendence()
	p.next_token()
	expr.Right = p.parse_expression(precedence)
	return expr
}

func (p *Parser) peek_precedence() int {
	if p, ok := precedences[p.peek_token.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) cur_precendence() int {
	if p, ok := precedences[p.cur_token.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parse_prefix_expression() ast.Expression {
	defer untrace(trace("parse_prefix_expression"))
	exp := &ast.Prefix_expression{Token: p.cur_token, Operator: p.cur_token.Literal}

	p.next_token()

	exp.Right = p.parse_expression(PREFIX)

	return exp
}

func (p *Parser) parse_integer_literal() ast.Expression {
	defer untrace(trace("parse_integer_literal"))
	lit := &ast.Integer_literal{Token: p.cur_token}

	value, err := strconv.ParseInt(p.cur_token.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.cur_token.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value

	return lit
}

func (p *Parser) parse_identifier() ast.Expression {
	return &ast.Identifier{Token: p.cur_token, Value: p.cur_token.Literal}

}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peek_token.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) next_token() {
	p.cur_token = p.peek_token
	p.peek_token = p.l.NextToken()
}

func (p *Parser) Parse_program() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.cur_token.Type != token.EOF {
		statement := p.parse_statement()

		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}
		p.next_token()
	}
	return program
}

func (p *Parser) parse_statement() ast.Statement {
	switch p.cur_token.Type {
	case token.LET:
		return p.parse_let_statement()
	case token.RETURN:
		return p.parse_return_statement()
	default:
		return p.parse_expression_statement()
	}

}

func (p *Parser) parse_expression_statement() ast.Statement {
	defer untrace(trace("parse_expression_statement"))
	statement := &ast.Expression_statement{Token: p.cur_token}

	statement.Expression = p.parse_expression(LOWEST)

	if p.peek_token_is(token.SEMICOLON) {
		p.next_token()
	}
	return statement
}

func (p *Parser) parse_expression(precedence int) ast.Expression {
	defer untrace(trace("parse_expression"))
	prefix := p.prefix_Parse_Fns[p.cur_token.Type]

	if prefix == nil {
		p.no_prefix_parse_fn_error(p.cur_token.Type)
		return nil
	}

	left_expr := prefix()

	for !p.peek_token_is(token.SEMICOLON) && precedence < p.peek_precedence() {
		infix := p.infix_Parse_Fns[p.peek_token.Type]
		if infix == nil {
			return left_expr
		}
		p.next_token()

		left_expr = infix(left_expr)
	}

	return left_expr
}

func (p *Parser) parse_return_statement() ast.Statement {
	statement := &ast.Return_statement{Token: p.cur_token}
	p.next_token()

	for !p.cur_token_is(token.SEMICOLON) {
		p.next_token()
	}
	return statement
}

func (p *Parser) parse_let_statement() ast.Statement {
	statement := &ast.Let_statement{Token: p.cur_token}

	if !p.expect_peek(token.IDENT) {
		return nil
	}

	statement.Name = &ast.Identifier{Token: p.cur_token, Value: p.cur_token.Literal}
	if !p.expect_peek(token.ASSIGN) {
		return nil
	}

	for !p.cur_token_is(token.SEMICOLON) {
		p.next_token()
	}
	return statement
}

func (p *Parser) cur_token_is(t token.TokenType) bool {
	return p.cur_token.Type == t
}

func (p *Parser) peek_token_is(t token.TokenType) bool {
	return p.peek_token.Type == t
}

func (p *Parser) expect_peek(t token.TokenType) bool {
	if p.peek_token_is(t) {
		p.next_token()
		return true
	} else {
		p.peekError(t)
		return false
	}

}

func (p *Parser) no_prefix_parse_fn_error(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) register_prefix(tokenType token.TokenType, fn prefix_Parse_Fn) {
	p.prefix_Parse_Fns[tokenType] = fn
}

func (p *Parser) register_infix(tokenType token.TokenType, fn infix_Parse_fn) {
	p.infix_Parse_Fns[tokenType] = fn
}
