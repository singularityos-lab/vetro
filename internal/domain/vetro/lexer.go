package vetro

// TokenType represents the type of a token.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenString
	TokenNumber
	TokenLParen
	TokenRParen
	TokenLBrace
	TokenRBrace
	TokenColon
	TokenDot
	TokenComma
	TokenILLEGAL
)

// Token represents a lexical token with source position for error reporting.
type Token struct {
	Type    TokenType
	Literal string
	Pos     int // absolute byte offset
	Line    int // 1-indexed line number
	Col     int // 1-indexed column number
}

// Lexer holds the state of the scanner.
type Lexer struct {
	input        string
	position     int  // current position in input
	readPosition int  // current reading position in input
	ch           byte // current char under examination
	line         int  // current line number
	col          int  // current column number
}

// NewLexer returns a new Lexer instance.
func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, line: 1, col: 1}
	l.readChar()
	return l
}

// readChar reads the next character and advances the position.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}

	if l.readPosition > 0 && l.readPosition <= len(l.input) && l.input[l.readPosition-1] == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}

	l.position = l.readPosition
	l.readPosition++
}

// skipWhitespace skips over whitespace characters and comments.
func (l *Lexer) skipWhitespace() {
	for {
		if l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			l.readChar()
		} else if l.ch == '/' && l.peekChar() == '*' {
			// Skip block comment /* ... */
			l.readChar() // skip '/'
			l.readChar() // skip '*'
			for l.ch != 0 {
				if l.ch == '*' && l.peekChar() == '/' {
					l.readChar() // skip '*'
					l.readChar() // skip '/'
					break
				}
				l.readChar()
			}
		} else if l.ch == '/' && l.peekChar() == '/' {
			// Skip line comment // ...
			l.readChar() // skip '/'
			l.readChar() // skip '/'
			for l.ch != 0 && l.ch != '\n' {
				l.readChar()
			}
		} else {
			break
		}
	}
}

// peekChar returns the next character without consuming.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	tok.Pos = l.position
	tok.Line = l.line
	tok.Col = l.col

	switch l.ch {
	case '(':
		tok.Type = TokenLParen
		tok.Literal = "("
		l.readChar()
	case ')':
		tok.Type = TokenRParen
		tok.Literal = ")"
		l.readChar()
	case '{':
		tok.Type = TokenLBrace
		tok.Literal = "{"
		l.readChar()
	case '}':
		tok.Type = TokenRBrace
		tok.Literal = "}"
		l.readChar()
	case ':':
		tok.Type = TokenColon
		tok.Literal = ":"
		l.readChar()
	case '.':
		tok.Type = TokenDot
		tok.Literal = "."
		l.readChar()
	case ',':
		tok.Type = TokenComma
		tok.Literal = ","
		l.readChar()
	case '"':
		tok.Type = TokenString
		l.readChar() // consume opening quote
		start := l.position
		for l.ch != '"' && l.ch != 0 {
			l.readChar()
		}
		tok.Literal = l.input[start:l.position]
		if l.ch == '"' {
			l.readChar() // consume closing quote
		}
	case 0:
		tok.Literal = ""
		tok.Type = TokenEOF
	default:
		if isLetter(l.ch) {
			start := l.position
			for isLetter(l.ch) || isDigit(l.ch) {
				l.readChar()
			}
			tok.Literal = l.input[start:l.position]
			tok.Type = TokenIdent
		} else if isDigit(l.ch) {
			start := l.position
			for isDigit(l.ch) {
				l.readChar()
			}
			tok.Literal = l.input[start:l.position]
			tok.Type = TokenNumber
		} else {
			tok.Type = TokenILLEGAL
			tok.Literal = string(l.ch)
			l.readChar()
		}
	}
	return tok
}

// isLetter returns true if the byte is a letter or underscore.
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit returns true if the byte is a digit.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// Tokenize runs the lexer and returns a slice of tokens.
func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return tokens, nil
}
