package markup

import (
	"strings"

	"github.com/frk/httptest/internal/markup/lexer"
)

func CSV(data []byte) string {
	var b strings.Builder

loop:
	for t := range lexer.CSV(data) {
		switch t.Type {
		case lexer.CSV_COM:
			b.Write([]byte(`<span class="token csv-com">,</span>`))
		case lexer.CSV_HED:
			b.Write([]byte(`<span class="token csv-header">`))
			b.Write(t.Value)
			b.Write([]byte(`</span>`))
		case lexer.CSV_FLD:
			b.Write([]byte(`<span class="token csv-field">`))
			b.Write(t.Value)
			b.Write([]byte(`</span>`))
		case lexer.CSV_NL:
			b.Write(t.Value)
		case lexer.CSV_EOF:
			break loop
		}
	}

	return b.String()
}
