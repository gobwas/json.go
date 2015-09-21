package js

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

type Type int

const (
	OBJECT_OPEN Type = iota
	OBJECT_CLOSE
	ARRAY_OPEN
	ARRAY_CLOSE
	STRING
	NUMBER
	COMMA
	COLON
	NULL
	TRUE
	FALSE
)

type Token struct {
	Type   Type
	Length int
}

type Parser struct {
	Index  int
	Offset int
	Data   []rune
	State  Type
}

var NumberStart = regexp.MustCompile("\\d")
var NumberBody = regexp.MustCompile("[0-9\\+\\-\\.eE]")
var WhiteSpace = regexp.MustCompile("[ \\t\\n\\r]")

func NewParser(input string) *Parser {
	runes := []rune(input)
	return &Parser{
		Offset: len(runes),
		Data:   runes,
	}
}

func Parse(input string) (interface{}, error) {
	parser := NewParser(input)
	return parser.Parse()
}

func (self *Parser) Parse() (interface{}, error) {
	return self.value()
}

func (self *Parser) value() (interface{}, error) {
	self.skipWhitespace()

	token, err := self.getToken()
	if err != nil {
		return nil, err
	}

	switch token.Type {
	case OBJECT_OPEN:
		return self.object()
	case ARRAY_OPEN:
		return self.array()
	case STRING:
		return self.string()
	case NUMBER:
		return self.number()
	case NULL:
		self.move(token.Length)
		return nil, nil
	case FALSE:
		self.move(token.Length)
		return false, nil
	case TRUE:
		self.move(token.Length)
		return true, nil
	}

	return nil, errors.New("Could not parse token")
}

func (self *Parser) object() (map[string]interface{}, error) {
	// skip first "{"
	self.move(1)

	obj := make(map[string]interface{})

	for {
		self.skipWhitespace()

		token, err := self.getToken()
		if err != nil {
			return nil, err
		}

		switch token.Type {
		case OBJECT_CLOSE:
			self.move(1)
			return obj, nil
		case COMMA:
			self.move(1)
		default:
			key, err := self.string()
			if err != nil {
				return nil, err
			}

			self.skipWhitespace()

			token, err := self.getToken()
			if err != nil {
				return nil, err
			}
			if token.Type != COLON {
				return nil, errors.New("Invalid JSON")
			}

			self.move(1)
			self.skipWhitespace()

			value, err := self.value()
			if err != nil {
				return nil, err
			}

			obj[key] = value
		}
	}

	return obj, nil
}

func (self *Parser) array() ([]interface{}, error) {
	// skip first "["
	self.move(1)

	array := make([]interface{}, 0)

	for {
		self.skipWhitespace()

		token, err := self.getToken()
		if err != nil {
			return nil, err
		}

		switch token.Type {
		case ARRAY_CLOSE:
			self.move(1)
			return array, nil
		case COMMA:
			self.move(1)
			continue
		}

		value, err := self.value()
		if err != nil {
			return nil, err
		}

		array = append(array, value)
	}

	return array, nil
}

func (self *Parser) string() (string, error) {
	var (
		str    string
		screen bool
		hex    bool
	)

	// skip first '"'
	self.move(1)

json:
	for ; !self.isComplete(); self.move(1) {
		symbol := self.Data[self.Index]

		if hex {
			hex = false

			codeStr := self.sliceCurrent(4, true)
			code, err := strconv.ParseUint(string(codeStr), 16, 64)
			if err != nil {
				return "", err
			}

			str += fmt.Sprintf("%c", code)

			continue
		}

		if screen {
			screen = false
			switch symbol {
			case '\\', '"', '/':
				str += string(symbol)
			case 'b', 'f', 'n', 'r', 't':
				str += `\` + string(symbol)
			case 'u':
				if self.Offset < 4 {
					return "", errors.New("Unexpected")
				}

				hex = true
			}

			continue
		}

		switch symbol {
		case '"':
			// skip last '"'
			self.move(1)
			break json
		case '\\':
			screen = true
			continue
		}

		str += string(symbol)
	}

	return str, nil
}

func (self *Parser) number() (float64, error) {
	var str string
	for ; !self.isComplete(); self.move(1) {
		symbol := string(self.Data[self.Index])
		if NumberBody.MatchString(symbol) {
			str += symbol
		} else {
			break
		}
	}

	number, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, err
	}

	return number, nil
}

func (self *Parser) getToken() (token *Token, err error) {
	switch self.Data[self.Index] {
	case '{':
		return &Token{OBJECT_OPEN, 1}, nil
	case '}':
		return &Token{OBJECT_CLOSE, 1}, nil
	case '[':
		return &Token{ARRAY_OPEN, 1}, nil
	case ']':
		return &Token{ARRAY_CLOSE, 1}, nil
	case '"':
		return &Token{STRING, 1}, nil
	case ',':
		return &Token{COMMA, 1}, nil
	case ':':
		return &Token{COLON, 1}, nil
	}

	if NumberStart.MatchString(string(self.Data[self.Index])) {
		return &Token{NUMBER, 1}, nil
	}

	if self.Offset >= 4 && string(self.sliceCurrent(4, false)) == "null" {
		return &Token{NULL, 4}, nil
	}

	if self.Offset >= 4 && string(self.sliceCurrent(4, false)) == "true" {
		return &Token{TRUE, 4}, nil
	}

	if self.Offset >= 5 && string(self.sliceCurrent(5, false)) == "false" {
		return &Token{FALSE, 5}, nil
	}

	return nil, errors.New("Unknown token")
}

func (self *Parser) skipWhitespace() {
	for ; !self.isComplete(); self.move(1) {
		if !WhiteSpace.MatchString(string(self.Data[self.Index])) {
			break
		}
	}
}

func (self *Parser) move(step int) {
	self.Index += step
	self.Offset -= step
}

func (self *Parser) sliceCurrent(len int, move bool) []rune {
	return self.slice(0, len, move)
}

func (self *Parser) slice(shift int, len int, move bool) []rune {
	ret := self.Data[self.Index+shift : self.Index+len+shift]
	if move {
		self.move(len + shift - 1)
	}

	return ret
}

func (self *Parser) isComplete() bool {
	return self.Offset == 0
}
