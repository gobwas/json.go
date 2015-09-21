package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode"
)

type Type int

const (
	// Special tokens
	ILLEGAL Type = iota
	EOF
	WHITESPACE

	// Literals
	NUMBER
	STRING

	// Delimiters
	CURLY_BRACE_OPEN
	CURLY_BRACE_CLOSE
	SQUARED_BRACE_OPEN
	SQUARED_BRACE_CLOSE
	COMMA
	COLON

	// Reserved words
	NULL
	TRUE
	FALSE
)

const eof = rune(0)

type Token struct {
	Type    Type
	Literal string
	Runes   []rune
}

func (self Token) String() string {
	switch self.Type {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case WHITESPACE:
		return "WHITESPACE"
	case NUMBER:
		return "NUMBER"
	case STRING:
		return "STRING"
	case CURLY_BRACE_OPEN:
		return "CURLY_BRACE_OPEN"
	case CURLY_BRACE_CLOSE:
		return "CURLY_BRACE_CLOSE"
	case SQUARED_BRACE_OPEN:
		return "SQUARED_BRACE_OPEN"
	case SQUARED_BRACE_CLOSE:
		return "SQUARED_BRACE_CLOSE"
	case COMMA:
		return "COMMA"
	case COLON:
		return "COLON"
	case NULL:
		return "NULL"
	case TRUE:
		return "TRUE"
	case FALSE:
		return "FALSE"
	}

	return ""
}

func isWhitespace(char rune) bool {
	return char == ' ' || char == '\n' || char == '\r' || char == '\t'
}

func isLowerCaseLetter(char rune) bool {
	return char >= 'a' && char <= 'z'
}

func isNumeric(char rune) bool {
	return (char >= '0' && char <= '9') || char == '-'
}

type Scanner struct {
	reader *bufio.Reader
}

// NewScanner returns a new instance of Scanner.
func NewScanner(reader io.Reader) *Scanner {
	return &Scanner{bufio.NewReader(reader)}
}

func (self *Scanner) read() rune {
	char, _, err := self.reader.ReadRune()
	if err != nil {
		return eof
	}
	return char
}

func (self *Scanner) unread() {
	self.reader.UnreadRune()
}

func (self *Scanner) Scan() (*Token, error) {
	// Read the next rune.
	char := self.read()

	switch {
	// If we see whitespace then consume all contiguous whitespace.
	case isWhitespace(char):
		self.unread()
		return self.scanWhitespace()
	// If we see a letter then consume as an ident or reserved word.
	case isLowerCaseLetter(char):
		self.unread()
		return self.scanIdentifier()
	case char == '"':
		self.unread()
		return self.scanString()
	case isNumeric(char):
		self.unread()
		return self.scanNumber()
	}

	// Otherwise read the individual character.
	switch char {
	case eof:
		return &Token{EOF, string(char), []rune{char}}, nil
	case '{':
		return &Token{CURLY_BRACE_OPEN, string(char), []rune{char}}, nil
	case '}':
		return &Token{CURLY_BRACE_CLOSE, string(char), []rune{char}}, nil
	case '[':
		return &Token{SQUARED_BRACE_OPEN, string(char), []rune{char}}, nil
	case ']':
		return &Token{SQUARED_BRACE_CLOSE, string(char), []rune{char}}, nil
	case ':':
		return &Token{COLON, string(char), []rune{char}}, nil
	case ',':
		return &Token{COMMA, string(char), []rune{char}}, nil
	}

	return &Token{ILLEGAL, string(char), []rune{char}}, nil
}

func (self *Scanner) scanWhitespace() (*Token, error) {
	var (
		buf  bytes.Buffer
		stop bool
	)

	for !stop {
		char := self.read()
		switch {
		case isWhitespace(char):
			buf.WriteRune(char)
		case char == eof:
			stop = true
		default:
			stop = true
			self.unread()
		}
	}

	return &Token{WHITESPACE, buf.String(), bytes.Runes(buf.Bytes())}, nil
}

type NumberState int

const (
	SIGN NumberState = iota
	DECIMAL_LESS_THAN_ONE
	DECIMAL
	FRACTIONAL
	EXPONENTIAL_SIGN
	EXPONENTIAL
)

func (self *Scanner) scanNumber() (*Token, error) {
	var (
		buf   bytes.Buffer
		stop  bool
		state NumberState
	)

	char := self.read()
	if !isNumeric(char) {
		self.unread()
		return &Token{NUMBER, buf.String(), bytes.Runes(buf.Bytes())}, nil
	}
	switch char {
	case '-':
		state = SIGN
	case '0':
		state = DECIMAL_LESS_THAN_ONE
	default:
		state = DECIMAL
	}
	buf.WriteRune(char)

	for !stop {
		char := self.read()

		switch state {
		case SIGN:
			buf.WriteRune(char)
			if char == '0' {
				state = DECIMAL_LESS_THAN_ONE
			} else {
				state = DECIMAL
			}
			continue
		case DECIMAL_LESS_THAN_ONE:
			if char == '.' {
				buf.WriteRune(char)
				state = FRACTIONAL
				continue
			}
		case DECIMAL:
			switch {
			case char == '.':
				buf.WriteRune(char)
				state = FRACTIONAL
				continue
			case char >= '0' && char <= '9':
				buf.WriteRune(char)
				continue
			}
		case FRACTIONAL:
			switch {
			case char == 'E' || char == 'e':
				buf.WriteRune(char)
				state = EXPONENTIAL_SIGN
				continue
			case char >= '0' && char <= '9':
				buf.WriteRune(char)
				continue
			}
		case EXPONENTIAL_SIGN:
			if char == '+' || char == '-' || (char >= '0' && char <= '9') {
				buf.WriteRune(char)
				state = EXPONENTIAL
				continue
			}
		case EXPONENTIAL:
			if char >= '0' && char <= '9' {
				buf.WriteRune(char)
				continue
			}
		}

		self.unread()
		stop = true
	}

	return &Token{NUMBER, buf.String(), bytes.Runes(buf.Bytes())}, nil
}

func (self *Scanner) scanString() (*Token, error) {
	var (
		buf    bytes.Buffer
		screen bool
		stop   bool
	)

	// skip first quote
	char := self.read()
	if char != '"' {
		// todo err here?
		self.unread()
		return &Token{STRING, buf.String(), bytes.Runes(buf.Bytes())}, nil
	}
	buf.WriteRune(char)

	for !stop {
		char := self.read()

		if unicode.IsControl(char) {
			self.unread()
			stop = true
			continue
		}

		if screen {
			screen = false

			switch char {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't', 'u':
				buf.WriteRune(char)
			default:
				return nil, fmt.Errorf("unexpected end of input: %q", string(char))
			}

			continue
		}

		switch {
		case char == '\\':
			screen = true
			buf.WriteRune(char)
		case char == '"':
			buf.WriteRune(char)
			stop = true
		case char == eof:
			stop = true
		default:
			buf.WriteRune(char)
		}
	}

	return &Token{STRING, buf.String(), bytes.Runes(buf.Bytes())}, nil
}

func (self *Scanner) scanIdentifier() (*Token, error) {
	var (
		buf  bytes.Buffer
		stop bool
	)

	for !stop {
		char := self.read()
		switch {
		case isLowerCaseLetter(char):
			buf.WriteRune(char)
		case char == eof:
			stop = true
		default:
			self.unread()
			stop = true
		}
	}

	lit := buf.String()
	switch lit {
	case "null":
		return &Token{NULL, lit, bytes.Runes(buf.Bytes())}, nil
	case "true":
		return &Token{TRUE, lit, bytes.Runes(buf.Bytes())}, nil
	case "false":
		return &Token{FALSE, lit, bytes.Runes(buf.Bytes())}, nil
	default:
		return &Token{ILLEGAL, lit, bytes.Runes(buf.Bytes())}, nil
	}
}
