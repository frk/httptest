package mimelexer

import (
	"testing"

	"github.com/frk/compare"
)

func TestJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []JSONToken
	}{{
		name:  "empty input",
		input: ``,
		want:  []JSONToken{{JSON_EOF, ""}},
	}, {
		name:  "empty string",
		input: `""`,
		want:  []JSONToken{{JSON_STR, `""`}, {JSON_EOF, ""}},
	}, {
		name:  "empty object",
		input: `{}`,
		want:  []JSONToken{{JSON_LCB, `{`}, {JSON_RCB, `}`}, {JSON_EOF, ""}},
	}, {
		name:  "empty array",
		input: `[]`,
		want:  []JSONToken{{JSON_LSB, `[`}, {JSON_RSB, `]`}, {JSON_EOF, ""}},
	}, {
		name:  "a string",
		input: `"foo bar"`,
		want:  []JSONToken{{JSON_STR, `"foo bar"`}, {JSON_EOF, ""}},
	}, {
		name:  "an array",
		input: `[123,"foo bar",true,null,{}]`,
		want: []JSONToken{
			{JSON_LSB, `[`},
			{JSON_NUM, `123`}, {JSON_COM, `,`},
			{JSON_STR, `"foo bar"`}, {JSON_COM, `,`},
			{JSON_TRU, `true`}, {JSON_COM, `,`},
			{JSON_NUL, `null`}, {JSON_COM, `,`},
			{JSON_LCB, `{`}, {JSON_RCB, `}`},
			{JSON_RSB, `]`},
			{JSON_EOF, ""},
		},
	}, {
		name: "an array with white space",
		input: `[
			123 ,
			"foo bar" ,
			true ,
			null ,
			{  }
		]`, //`
		want: []JSONToken{
			{JSON_LSB, `[`},
			{JSON_WS, "\n\t\t\t"}, {JSON_NUM, `123`}, {JSON_WS, " "}, {JSON_COM, `,`},
			{JSON_WS, "\n\t\t\t"}, {JSON_STR, `"foo bar"`}, {JSON_WS, " "}, {JSON_COM, `,`},
			{JSON_WS, "\n\t\t\t"}, {JSON_TRU, `true`}, {JSON_WS, " "}, {JSON_COM, `,`},
			{JSON_WS, "\n\t\t\t"}, {JSON_NUL, `null`}, {JSON_WS, " "}, {JSON_COM, `,`},
			{JSON_WS, "\n\t\t\t"}, {JSON_LCB, `{`}, {JSON_WS, "  "}, {JSON_RCB, `}`},
			{JSON_WS, "\n\t\t"}, {JSON_RSB, `]`},
			{JSON_EOF, ""},
		},
	}, {
		name:  "an object",
		input: `{"num":-0.123,"text":"foo bar","list":[{"key1":true},{"key2":null}]}`,
		want: []JSONToken{
			{JSON_LCB, `{`},
			{JSON_KEY, `"num"`}, {JSON_CLN, `:`}, {JSON_NUM, `-0.123`}, {JSON_COM, `,`},
			{JSON_KEY, `"text"`}, {JSON_CLN, `:`}, {JSON_STR, `"foo bar"`}, {JSON_COM, `,`},
			{JSON_KEY, `"list"`}, {JSON_CLN, `:`},
			{JSON_LSB, `[`}, {JSON_LCB, `{`},
			{JSON_KEY, `"key1"`}, {JSON_CLN, `:`}, {JSON_TRU, `true`},
			{JSON_RCB, `}`}, {JSON_COM, `,`}, {JSON_LCB, `{`},
			{JSON_KEY, `"key2"`}, {JSON_CLN, `:`}, {JSON_NUL, `null`},
			{JSON_RCB, `}`}, {JSON_RSB, `]`},
			{JSON_RCB, `}`},
			{JSON_EOF, ""},
		},
	}, {
		name: "an object with white space",
		input: `{
			"num": -0.123,
			"text": "foo bar",
			"list": [
				{ "key1": true },

				{
					"key2": null
				}
			]
		}`, //`
		want: []JSONToken{
			0:  {JSON_LCB, `{`},
			1:  {JSON_WS, "\n\t\t\t"},
			2:  {JSON_KEY, `"num"`},
			3:  {JSON_CLN, `:`},
			4:  {JSON_WS, " "},
			5:  {JSON_NUM, `-0.123`},
			6:  {JSON_COM, `,`},
			7:  {JSON_WS, "\n\t\t\t"},
			8:  {JSON_KEY, `"text"`},
			9:  {JSON_CLN, `:`},
			10: {JSON_WS, " "},
			11: {JSON_STR, `"foo bar"`},
			12: {JSON_COM, `,`},
			13: {JSON_WS, "\n\t\t\t"},
			14: {JSON_KEY, `"list"`},
			15: {JSON_CLN, `:`},
			16: {JSON_WS, " "},
			17: {JSON_LSB, `[`},
			18: {JSON_WS, "\n\t\t\t\t"},
			19: {JSON_LCB, `{`},
			20: {JSON_WS, " "},
			21: {JSON_KEY, `"key1"`},
			22: {JSON_CLN, `:`},
			23: {JSON_WS, " "},
			24: {JSON_TRU, `true`},
			25: {JSON_WS, " "},
			26: {JSON_RCB, `}`},
			27: {JSON_COM, `,`},
			28: {JSON_WS, "\n\n\t\t\t\t"},
			29: {JSON_LCB, `{`},
			30: {JSON_WS, "\n\t\t\t\t\t"},
			31: {JSON_KEY, `"key2"`},
			32: {JSON_CLN, `:`},
			33: {JSON_WS, " "},
			34: {JSON_NUL, `null`},
			35: {JSON_WS, "\n\t\t\t\t"},
			36: {JSON_RCB, `}`},
			37: {JSON_WS, "\n\t\t\t"},
			38: {JSON_RSB, `]`},
			39: {JSON_WS, "\n\t\t"},
			40: {JSON_RCB, `}`},
			41: {JSON_EOF, ""},
		},
	}}

	var agg_tokens = func(input string) (tokens []JSONToken) {
		ch := JSON([]byte(input))
		for {
			tok := <-ch
			tokens = append(tokens, tok)
			if tok.Type == JSON_EOF {
				break
			}
		}
		return tokens
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agg_tokens(tt.input)
			if e := compare.Compare(got, tt.want); e != nil {
				t.Error(e)
			}
		})
	}
}
