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
	token.LPAREN:   CALL,
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
	p.register_prefix(token.TRUE, p.parse_boolean)
	p.register_prefix(token.FALSE, p.parse_boolean)
	p.register_prefix(token.LPAREN, p.parse_grouped_expression)
	p.register_prefix(token.IF, p.parse_if_expression)
	p.register_prefix(token.FUNCTION, p.parse_function_literal)

	p.register_infix(token.PLUS, p.parse_infix_expression)
	p.register_infix(token.MINUS, p.parse_infix_expression)
	p.register_infix(token.SLASH, p.parse_infix_expression)
	p.register_infix(token.ASTERISK, p.parse_infix_expression)
	p.register_infix(token.EQ, p.parse_infix_expression)
	p.register_infix(token.NOT_EQ, p.parse_infix_expression)
	p.register_infix(token.LT, p.parse_infix_expression)
	p.register_infix(token.GT, p.parse_infix_expression)
	p.register_infix(token.LPAREN, p.parse_call_expression)

	return p
}

func (p *Parser) parse_call_expression(function ast.Expression) ast.Expression {
	expr := &ast.Call_expression{Token: p.cur_token, Function: function}
	expr.Arguments = p.parse_call_arguments()
	return expr
}

func (p *Parser) parse_call_arguments() []ast.Expression {
	arguments := []ast.Expression{}

	if p.peek_token_is(token.RPAREN) {
		p.next_token()
		return arguments
	}
	p.next_token()
	arguments = append(arguments, p.parse_expression(LOWEST))

	for p.peek_token_is(token.COMMA) {
		p.next_token()
		p.next_token()
		arguments = append(arguments, p.parse_expression(LOWEST))

	}
	if !p.expect_peek(token.RPAREN) {
		return nil
	}
	return arguments
}

func (p *Parser) parse_function_literal() ast.Expression {
	expr := &ast.Function_literal{Token: p.cur_token}
	if !p.expect_peek(token.LPAREN) {
		return nil
	}

	expr.Parameters = p.parse_function_parameters()

	if !p.expect_peek(token.LBRACE) {
		return nil
	}

	expr.Body = p.parse_block_statement()

	return expr
}

func (p *Parser) parse_function_parameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peek_token_is(token.RPAREN) {
		p.next_token()
		return identifiers
	}
	p.next_token()
	ident := &ast.Identifier{Token: p.cur_token, Value: p.cur_token.Literal}
	identifiers = append(identifiers, ident)

	for p.peek_token_is(token.COMMA) {
		p.next_token()
		p.next_token()
		ident := &ast.Identifier{Token: p.cur_token, Value: p.cur_token.Literal}
		identifiers = append(identifiers, ident)
	}
	if !p.expect_peek(token.RPAREN) {
		return nil
	}

	return identifiers

}

func (p *Parser) parse_if_expression() ast.Expression {
	expr := &ast.If_expression{Token: p.cur_token}

	if !p.expect_peek(token.LPAREN) {
		return nil
	}

	p.next_token()

	expr.Condition = p.parse_expression(LOWEST)

	if !p.expect_peek(token.RPAREN) {
		return nil
	}

	if !p.expect_peek(token.LBRACE) {
		return nil
	}

	expr.Consequence = p.parse_block_statement()

	if p.peek_token_is(token.ELSE) {
		p.next_token()

		if !p.expect_peek(token.LBRACE) {
			return nil
		}
		expr.Alternative = p.parse_block_statement()
	}

	return expr

}

func (p *Parser) parse_block_statement() *ast.Block_statement {
	expr := &ast.Block_statement{Token: p.cur_token}
	expr.Statements = []ast.Statement{}

	p.next_token()

	for !p.cur_token_is(token.RBRACE) && !p.cur_token_is(token.EOF) {
		stmt := p.parse_statement()
		if stmt != nil {
			expr.Statements = append(expr.Statements, stmt)
		}
		p.next_token()

	}
	return expr

}

func (p *Parser) parse_grouped_expression() ast.Expression {
	p.next_token()

	expr := p.parse_expression(LOWEST)

	if !p.expect_peek(token.RPAREN) {
		return nil
	}
	return expr
}

func (p *Parser) parse_boolean() ast.Expression {
	return &ast.Boolean{Token: p.cur_token, Value: p.cur_token_is(token.TRUE)}
}

func (p *Parser) parse_infix_expression(left ast.Expression) ast.Expression {
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
	exp := &ast.Prefix_expression{Token: p.cur_token, Operator: p.cur_token.Literal}

	p.next_token()

	exp.Right = p.parse_expression(PREFIX)

	return exp
}

func (p *Parser) parse_integer_literal() ast.Expression {
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
	statement := &ast.Expression_statement{Token: p.cur_token}

	statement.Expression = p.parse_expression(LOWEST)

	if p.peek_token_is(token.SEMICOLON) {
		p.next_token()
	}
	return statement
}

func (p *Parser) parse_expression(precedence int) ast.Expression {
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

	statement.Return_value = p.parse_expression(LOWEST)

	if p.peek_token_is(token.SEMICOLON) {
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
	p.next_token()

	statement.Value = p.parse_expression(LOWEST)
	if p.peek_token_is(token.SEMICOLON) {
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
