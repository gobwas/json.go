package parser

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

func ParseString(str string) (interface{}, error) {
	parser := NewParser(strings.NewReader(str))
	return parser.Parse()
}

type Parser struct {
	scanner *Scanner
	buf     struct {
		token *Token
		size  int
	}
}

func NewParser(reader io.Reader) *Parser {
	return &Parser{scanner: NewScanner(reader)}
}

func (self *Parser) scan() (*Token, error) {
	if self.buf.size == 1 {
		self.buf.size = 0
		return self.buf.token, nil
	}

	token, err := self.scanner.Scan()
	if err != nil {
		return nil, err
	}
	self.buf.token = token

	return token, nil
}

func (self *Parser) unscan() {
	self.buf.size = 1
}

func (self *Parser) scanIgnoreWhitespace() (*Token, error) {
	token, err := self.scan()
	if err != nil {
		return nil, err
	}
	if token.Type == WHITESPACE {
		return self.scan()
	}

	return token, nil
}

func parseString(runes []rune) (string, error) {
	var (
		str    string
		screen bool
		hex    bool
		stop   bool
	)

	// skip first quote
	index := 1
	length := len(runes)

	for ; index < length && !stop; index++ {
		symbol := runes[index]

		if hex {
			hex = false

			codeStr := string(runes[index : index+4])
			index += 3
			code, err := strconv.ParseUint(codeStr, 16, 64)
			if err != nil {
				return "", fmt.Errorf("could not parse hex charcode: %s", err)
			}

			str += fmt.Sprintf("%c", code)

			continue
		}

		if screen {
			screen = false

			switch symbol {
			case '\\', '"', '/':
				str += string(symbol)
			case 'b':
				str += "\b"
			case 'f':
				str += "\f"
			case 'n':
				str += "\n"
			case 'r':
				str += "\r"
			case 't':
				str += "\t"
			case 'u':
				if length-index < 4 {
					return "", fmt.Errorf("unexpected length of hex symbol")
				}

				hex = true
			default:
				return "", fmt.Errorf("unexpected token after screen character: %q", string(symbol))
			}

			continue
		}

		switch symbol {
		case '"':
			stop = true
			continue
		case '\\':
			screen = true
			continue
		}

		str += string(symbol)
	}

	return str, nil
}

func parseNumber(runes []rune) (float64, error) {
	if number, err := strconv.ParseFloat(string(runes), 64); err != nil {
		return 0, fmt.Errorf("could not parse number: %s", err)
	} else {
		return number, nil
	}
}

func (self *Parser) scanValue() (interface{}, error) {
	token, err := self.scanIgnoreWhitespace()
	if err != nil {
		return nil, err
	}

	switch token.Type {
	case CURLY_BRACE_OPEN:
		self.unscan()
		return self.scanObject()
	case SQUARED_BRACE_OPEN:
		self.unscan()
		return self.scanArray()
	case STRING:
		return parseString(token.Runes)
	case NUMBER:
		return parseNumber(token.Runes)
	case NULL:
		return nil, nil
	case FALSE:
		return false, nil
	case TRUE:
		return true, nil
	case ILLEGAL:
		return nil, fmt.Errorf("illegal literal: %q", token.Literal)
	default:
		return nil, fmt.Errorf("could not parse token as value: %q", token.Literal)
	}
}

func (self *Parser) scanArray() ([]interface{}, error) {
	token, err := self.scan()
	if err != nil {
		return nil, err
	}
	if token.Type != SQUARED_BRACE_OPEN {
		return nil, fmt.Errorf("found %s %q, expected %q", token.String(), token.Literal, '[')
	}

	array := make([]interface{}, 0)

	for {
		token, err := self.scanIgnoreWhitespace()
		if err != nil {
			return nil, err
		}

		switch token.Type {
		case STRING, NUMBER, CURLY_BRACE_OPEN, SQUARED_BRACE_OPEN, TRUE, FALSE, NULL:
			self.unscan()
			value, err := self.scanValue()
			if err != nil {
				return nil, err
			}
			array = append(array, value)
		case SQUARED_BRACE_CLOSE:
			return array, nil
		case COMMA:
			continue
		}
	}

	return array, nil
}

func (self *Parser) scanObject() (map[string]interface{}, error) {
	token, err := self.scan()
	if err != nil {
		return nil, err
	}
	if token.Type != CURLY_BRACE_OPEN {
		return nil, fmt.Errorf("found %s %q, expected %q", token.String(), token.Literal, '{')
	}

	obj := make(map[string]interface{})

	for {
		token, err := self.scanIgnoreWhitespace()
		if err != nil {
			return nil, err
		}

		switch token.Type {
		case CURLY_BRACE_CLOSE:
			return obj, nil
		case COMMA:
			continue
		case STRING:
			key, err := parseString(token.Runes)
			if err != nil {
				return nil, err
			}

			token, err := self.scanIgnoreWhitespace()
			if err != nil {
				return nil, err
			}
			if token.Type != COLON {
				return nil, fmt.Errorf("found %s %q expected %q", token.String(), token.Literal, ':')
			}

			value, err := self.scanValue()
			if err != nil {
				return nil, err
			}

			obj[key] = value
		default:
			return nil, fmt.Errorf("unexpected token: %s %q", token.String(), token.Literal)
		}
	}

	return obj, nil
}

func (self *Parser) Parse() (interface{}, error) {
	token, err := self.scan()
	if err != nil {
		return nil, err
	}

	switch token.Type {
	case CURLY_BRACE_OPEN:
		self.unscan()
		return self.scanObject()
	case SQUARED_BRACE_OPEN:
		self.unscan()
		return self.scanArray()
	default:
		return nil, fmt.Errorf("found %s %q expected %q or %q", token.String(), token.Literal, '{', '[')
	}
}
