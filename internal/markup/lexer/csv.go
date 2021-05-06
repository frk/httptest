package lexer

type CSVTokenType uint8

const (
	_ CSVTokenType = iota

	CSV_COM // ','
	CSV_HED // header name
	CSV_FLD // record field
	CSV_NL  // new line
	CSV_EOF // end of the csv input
)

// mainly for debugging
func (t CSVTokenType) String() string { return csvTokenTypes[t] }

var csvTokenTypes = [...]string{
	CSV_COM: `','`,
	CSV_HED: `header`,
	CSV_FLD: `field`,
	CSV_NL:  `<nl>`,
	CSV_EOF: `<eof>`,
}

type CSVToken struct {
	Type  CSVTokenType
	Value []byte
}

func CSV(input []byte) (tokens <-chan CSVToken) {
	lx := new(csvlexer)
	lx.input = input
	lx.tokens = make(chan CSVToken)
	lx.head = true
	go lx.run()

	return lx.tokens
}

type csvlexer struct {
	input  []byte
	start  int
	pos    int
	width  int
	head   bool
	tokens chan CSVToken
}

func (lx *csvlexer) run() {
	for next := lx.inputStart; next != nil; {
		next = next()
	}
	close(lx.tokens)
}

func (lx *csvlexer) inputStart() stateFn {
	if lx.next() == eof {
		return lx.endOfFile
	}
	lx.backup()
	return lx.fieldStart
}

func (lx *csvlexer) fieldStart() stateFn {
	switch r := lx.next(); r {
	case eof, ',', '\n':
		return lx.emptyField
	case '"':
		return lx.escapedField
	}
	return lx.nonEscapedField
}

func (lx *csvlexer) endOfFile() stateFn {
	lx.emit(CSV_EOF)
	return nil
}

// position is known to be at ','
func (lx *csvlexer) emptyField() stateFn {
	lx.backup()
	lx.emit(CSV_FLD)
	return lx.fieldEnd
}

// position is known to be at opening '"'
func (lx *csvlexer) escapedField() stateFn {
loop:
	for {
		switch lx.next() {
		case eof: // shouldn't happen
			lx.ignore()
			return lx.endOfFile
		case '"':
			if lx.next() != '"' {
				lx.backup()
				lx.emit(CSV_FLD)
				break loop
			}
		}
	}
	return lx.fieldEnd
}

// position is known to be at beginning of a non-escaped field
func (lx *csvlexer) nonEscapedField() stateFn {
loop:
	for {
		switch lx.next() {
		case eof, ',', '\n':
			lx.backup()
			lx.emit(CSV_FLD)
			break loop
		}
	}
	return lx.fieldEnd
}

// position is known to be at a field's end
func (lx *csvlexer) fieldEnd() stateFn {
	switch lx.next() {
	case eof:
		return lx.endOfFile
	case ',':
		lx.emit(CSV_COM)
	case '\n':
		lx.emit(CSV_NL)
	}
	return lx.fieldStart
}

////////////////////////////////////////////////////////////////////////////////
// base commands
////////////////////////////////////////////////////////////////////////////////

// emit passes an item back to the client.
func (lx *csvlexer) emit(typ CSVTokenType) {
	switch typ {
	case CSV_NL:
		lx.head = false
	case CSV_FLD:
		if lx.head {
			typ = CSV_HED
		}
	}

	lx.tokens <- CSVToken{Type: typ, Value: lx.input[lx.start:lx.pos]}
	lx.start = lx.pos
}

// next returns the next byte in the input.
func (lx *csvlexer) next() (c byte) {
	if lx.pos >= len(lx.input) {
		lx.width = 0
		return eof
	}
	c = lx.input[lx.pos]
	lx.pos += 1
	lx.width = 1
	return c
}

// backup steps back one byte. Can bakcup only once per call to next.
func (lx *csvlexer) backup() {
	lx.pos -= lx.width
	lx.width = 0
}

// ignore skips over the pending input before this point.
func (lx *csvlexer) ignore() {
	lx.start = lx.pos
}
