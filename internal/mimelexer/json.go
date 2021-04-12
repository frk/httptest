package mimelexer

import (
	"strings"
	"unicode/utf8"
)

type JSON_TOKEN_TYPE uint16

const (
	_ JSON_TOKEN_TYPE = iota

	JSON_LSB // '['
	JSON_RSB // ']'
	JSON_LCB // '{'
	JSON_RCB // '}'
	JSON_CLN // ':'
	JSON_COM // ','
	JSON_STR // a json string
	JSON_NUM // a json number
	JSON_TRU // 'true'
	JSON_FAL // 'false'
	JSON_NUL // 'null'
	JSON_KEY // a json object key
	JSON_WS  // white space
	JSON_EOF // end of the json input
)

var jsonTokenTypes = [...]string{
	JSON_LSB: `'['`,
	JSON_RSB: `']'`,
	JSON_LCB: `'{'`,
	JSON_RCB: `'}'`,
	JSON_CLN: `':'`,
	JSON_COM: `','`,
	JSON_STR: `string`,
	JSON_NUM: `number`,
	JSON_TRU: `'true'`,
	JSON_FAL: `'false'`,
	JSON_NUL: `'null'`,
	JSON_KEY: `<obj_key>`,
	JSON_WS:  `<ws>`,
	JSON_EOF: `<eof>`,
}

// mainly for debugging
func (t JSON_TOKEN_TYPE) String() string { return jsonTokenTypes[t] }

type JSONToken struct {
	Type  JSON_TOKEN_TYPE
	Value string
}

func JSON(data []byte) (tokens <-chan JSONToken) {
	lx := new(jsonlexer)
	lx.input = string(data)
	lx.tokens = make(chan JSONToken)
	go lx.run()

	return lx.tokens
}

type jsonlexer struct {
	input  string
	start  int
	pos    int
	width  int
	nest   []rune
	tokens chan JSONToken
}

type stateFn func() stateFn

func (lx *jsonlexer) run() {
	for next := lx.valueStart; next != nil; {
		next = next()
	}
	close(lx.tokens)
}

////////////////////////////////////////////////////////////////////////////////
//
////////////////////////////////////////////////////////////////////////////////

func (lx *jsonlexer) valueStart() stateFn {
	lx.eatws()
	switch r := lx.next(); r {
	case eof:
		return lx.endOfFile
	case '{':
		return lx.objectStart
	case '[':
		return lx.arrayStart
	case '"':
		return lx.stringValue
	case 't':
		return lx.trueLiteral
	case 'f':
		return lx.falseLiteral
	case 'n':
		return lx.nullLiteral
	}
	return lx.numberValue
}

func (lx *jsonlexer) valueEnd() stateFn {
	lx.eatws()
	switch r := lx.next(); r {
	case eof:
		return lx.endOfFile
	case '}':
		return lx.objectEnd
	case ']':
		return lx.arrayEnd
	case ',':
		lx.emit(JSON_COM)

		// check if inside object
		if lx.nest[len(lx.nest)-1] == '{' {
			lx.eatws()
			if r := lx.next(); r == '"' {
				return lx.objectKey
			}
			return lx.objectEnd
		}
	}

	return lx.valueStart
}

// position is known to be at 'n'
func (lx *jsonlexer) nullLiteral() stateFn {
	lx.pos += 3 // len(`ull`)
	lx.emit(JSON_NUL)
	return lx.valueEnd
}

// position is known to be at 'f'
func (lx *jsonlexer) falseLiteral() stateFn {
	lx.pos += 4 // len(`alse`)
	lx.emit(JSON_FAL)
	return lx.valueEnd
}

// position is known to be at 't'
func (lx *jsonlexer) trueLiteral() stateFn {
	lx.pos += 3 // len(`rue`)
	lx.emit(JSON_TRU)
	return lx.valueEnd
}

// position is known to be at first char of the number
func (lx *jsonlexer) numberValue() stateFn {
	lx.acceptrun("0123456789eE+-.")
	lx.emit(JSON_NUM)
	return lx.valueEnd
}

// position is known to be at opening '"'
func (lx *jsonlexer) stringValue() stateFn {
loop:
	for esc := false; ; {
		switch r := lx.next(); {
		case r == '"' && !esc:
			break loop
		case r == '\\':
			esc = !esc
		default:
			esc = false
		}
	}

	// position at closing '"'
	lx.emit(JSON_STR)
	return lx.valueEnd
}

// position is known to be at opening '{'
func (lx *jsonlexer) objectStart() stateFn {
	lx.nest = append(lx.nest, '{')
	lx.emit(JSON_LCB)

	lx.eatws()
	if r := lx.next(); r == '}' {
		return lx.objectEnd
	}
	return lx.objectKey
}

// position is known to be at closing '}'
func (lx *jsonlexer) objectEnd() stateFn {
	lx.nest = lx.nest[:len(lx.nest)-1]
	lx.emit(JSON_RCB)
	return lx.valueEnd
}

// position is known to be at opening '"'
func (lx *jsonlexer) objectKey() stateFn {
loop:
	for esc := false; ; {
		switch r := lx.next(); {
		case r == '"' && !esc:
			break loop
		case r == '\\':
			esc = !esc
		default:
			esc = false
		}
	}
	lx.emit(JSON_KEY)

	lx.eatws()
	lx.next() // known to be ":"
	lx.emit(JSON_CLN)

	return lx.valueStart
}

// position is known to be at opening '['
func (lx *jsonlexer) arrayStart() stateFn {
	lx.nest = append(lx.nest, '[')
	lx.emit(JSON_LSB)

	lx.eatws()
	if r := lx.next(); r == ']' {
		return lx.arrayEnd
	}
	lx.backup()
	return lx.valueStart
}

// position is known to be at closing ']'
func (lx *jsonlexer) arrayEnd() stateFn {
	lx.nest = lx.nest[:len(lx.nest)-1]
	lx.emit(JSON_RSB)
	return lx.valueEnd
}

func (lx *jsonlexer) endOfFile() stateFn {
	lx.emit(JSON_EOF)
	return nil
}

////////////////////////////////////////////////////////////////////////////////
//
////////////////////////////////////////////////////////////////////////////////

// emit passes an item back to the client.
func (lx *jsonlexer) emit(typ JSON_TOKEN_TYPE) {
	lx.tokens <- JSONToken{Type: typ, Value: lx.input[lx.start:lx.pos]}
	lx.start = lx.pos
}

// next returns the next rune in the input.
func (lx *jsonlexer) next() (r rune) {
	if lx.pos >= len(lx.input) {
		lx.width = 0
		return eof
	}
	r, lx.width = utf8.DecodeRuneInString(lx.input[lx.pos:])
	lx.pos += lx.width
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (lx *jsonlexer) backup() {
	lx.pos -= lx.width
	lx.width = 0
}

// eatws consumes and emits a run of white space runes.
func (lx *jsonlexer) eatws() {
	if lx.acceptrun(" \t\r\n") > 0 {
		lx.emit(JSON_WS)
	}
}

// acceptrun consumes a run of runes from the valid set and returns the number consumed.
func (lx *jsonlexer) acceptrun(valid string) (num int) {
	for strings.ContainsRune(valid, lx.next()) {
		num += 1
	}
	lx.backup()
	return num
}
