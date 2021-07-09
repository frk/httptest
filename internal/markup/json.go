package markup

import (
	"html"
	"strings"

	"github.com/frk/httptest/internal/markup/lexer"
)

func JSON(data []byte) string {
	var b strings.Builder

loop:
	for t := range lexer.JSON(data) {
		switch t.Type {
		case lexer.JSON_LSB:
			b.Write([]byte(`<span class="token json-lsb">[</span>`))
		case lexer.JSON_RSB:
			b.Write([]byte(`<span class="token json-rsb">]</span>`))
		case lexer.JSON_LCB:
			b.Write([]byte(`<span class="token json-lcb">{</span>`))
		case lexer.JSON_RCB:
			b.Write([]byte(`<span class="token json-rcb">}</span>`))
		case lexer.JSON_CLN:
			b.Write([]byte(`<span class="token json-cln">:</span>`))
		case lexer.JSON_COM:
			b.Write([]byte(`<span class="token json-com">,</span>`))
		case lexer.JSON_STR:
			b.Write([]byte(`<span class="token json-str">`))
			b.Write([]byte(html.EscapeString(string(t.Value))))
			b.Write([]byte(`</span>`))
		case lexer.JSON_NUM:
			b.Write([]byte(`<span class="token json-num">`))
			b.Write(t.Value)
			b.Write([]byte(`</span>`))
		case lexer.JSON_TRU:
			b.Write([]byte(`<span class="token json-tru">true</span>`))
		case lexer.JSON_FAL:
			b.Write([]byte(`<span class="token json-fal">false</span>`))
		case lexer.JSON_NUL:
			b.Write([]byte(`<span class="token json-nul">null</span>`))
		case lexer.JSON_KEY:
			b.Write([]byte(`<span class="token json-key-q">"</span>`)) // key quote

			b.Write([]byte(`<span class="token json-key-t">`)) // key text
			b.Write([]byte(html.EscapeString(string(t.Value[1 : len(t.Value)-1]))))
			b.Write([]byte(`</span>`))

			b.Write([]byte(`<span class="token json-key-q">"</span>`)) // key quote
		case lexer.JSON_WS:
			b.Write(t.Value)
		case lexer.JSON_EOF:
			break loop
		}
	}

	return b.String()
}
