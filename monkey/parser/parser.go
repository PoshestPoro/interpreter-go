package parser

import (
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
)

type Parser struct {
	l *lexer.Lexer

	cur_token  token.Token
	peek_token token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}

	p.next_token()
	p.next_token()

	return p
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
	default:
		return nil
	}

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
		return false
	}

}
