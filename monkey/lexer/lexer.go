package lexer

import "monkey/token"

type Lexer struct {
	input         string
	position      int
	read_position int
	ch            byte
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.read_char()
	return l
}

func (l *Lexer) read_char() {
	if l.read_position >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.read_position]
	}
	l.position = l.read_position
	l.read_position += 1
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skip_whitespace()

	switch l.ch {
	case '=':
		if l.peek_char() == '=' {
			ch := l.ch
			l.read_char()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = new_token(token.ASSIGN, l.ch)
		}
	case '+':
		tok = new_token(token.PLUS, l.ch)
	case '-':
		tok = new_token(token.MINUS, l.ch)
	case '!':
		if l.peek_char() == '=' {
			ch := l.ch
			l.read_char()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = new_token(token.BANG, l.ch)
		}
	case '/':
		tok = new_token(token.SLASH, l.ch)
	case '*':
		tok = new_token(token.ASTERISK, l.ch)
	case '<':
		tok = new_token(token.LT, l.ch)
	case '>':
		tok = new_token(token.GT, l.ch)
	case ';':
		tok = new_token(token.SEMICOLON, l.ch)
	case '(':
		tok = new_token(token.LPAREN, l.ch)
	case ')':
		tok = new_token(token.RPAREN, l.ch)
	case ',':
		tok = new_token(token.COMMA, l.ch)
	case '{':
		tok = new_token(token.LBRACE, l.ch)
	case '}':
		tok = new_token(token.RBRACE, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if is_letter(l.ch) {
			tok.Literal = l.read_identifier()
			tok.Type = token.Lookup_identifier(tok.Literal)
			return tok
		} else if is_digit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.read_number()
			return tok
		} else {
			tok = new_token(token.ILLEGAL, l.ch)
		}
	}
	l.read_char()
	return tok
}

func (l *Lexer) read_identifier() string {
	position := l.position
	for is_letter(l.ch) {
		l.read_char()
	}
	return l.input[position:l.position]
}

func (l *Lexer) skip_whitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.read_char()
	}
}

func is_letter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func (l *Lexer) read_number() string {
	position := l.position
	for is_digit(l.ch) {
		l.read_char()
	}
	return l.input[position:l.position]
}

func (l *Lexer) peek_char() byte {
	if l.read_position >= len(l.input) {
		return 0
	} else {
		return l.input[l.read_position]
	}
}

func is_digit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
func new_token(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}
