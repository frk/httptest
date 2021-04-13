package lexer

import (
	"testing"

	"github.com/frk/compare"
)

func TestJSON(t *testing.T) {
	// helper
	B := func(s string) []byte { return []byte(s) }

	tests := []struct {
		name  string
		input string
		want  []JSONToken
	}{{
		name:  "empty input",
		input: ``,
		want:  []JSONToken{{JSON_EOF, B("")}},
	}, {
		name:  "empty string",
		input: `""`,
		want:  []JSONToken{{JSON_STR, B(`""`)}, {JSON_EOF, B("")}},
	}, {
		name:  "empty object",
		input: `{}`,
		want:  []JSONToken{{JSON_LCB, B(`{`)}, {JSON_RCB, B(`}`)}, {JSON_EOF, B("")}},
	}, {
		name:  "empty array",
		input: `[]`,
		want:  []JSONToken{{JSON_LSB, B(`[`)}, {JSON_RSB, B(`]`)}, {JSON_EOF, B("")}},
	}, {
		name:  "a string",
		input: `"foo bar"`,
		want:  []JSONToken{{JSON_STR, B(`"foo bar"`)}, {JSON_EOF, B("")}},
	}, {
		name:  "an array",
		input: `[123,"foo bar",true,null,{}]`,
		want: []JSONToken{
			{JSON_LSB, B(`[`)},
			{JSON_NUM, B(`123`)}, {JSON_COM, B(`,`)},
			{JSON_STR, B(`"foo bar"`)}, {JSON_COM, B(`,`)},
			{JSON_TRU, B(`true`)}, {JSON_COM, B(`,`)},
			{JSON_NUL, B(`null`)}, {JSON_COM, B(`,`)},
			{JSON_LCB, B(`{`)}, {JSON_RCB, B(`}`)},
			{JSON_RSB, B(`]`)},
			{JSON_EOF, B("")},
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
			{JSON_LSB, B(`[`)},
			{JSON_WS, B("\n\t\t\t")}, {JSON_NUM, B(`123`)}, {JSON_WS, B(" ")}, {JSON_COM, B(`,`)},
			{JSON_WS, B("\n\t\t\t")}, {JSON_STR, B(`"foo bar"`)}, {JSON_WS, B(" ")}, {JSON_COM, B(`,`)},
			{JSON_WS, B("\n\t\t\t")}, {JSON_TRU, B(`true`)}, {JSON_WS, B(" ")}, {JSON_COM, B(`,`)},
			{JSON_WS, B("\n\t\t\t")}, {JSON_NUL, B(`null`)}, {JSON_WS, B(" ")}, {JSON_COM, B(`,`)},
			{JSON_WS, B("\n\t\t\t")}, {JSON_LCB, B(`{`)}, {JSON_WS, B("  ")}, {JSON_RCB, B(`}`)},
			{JSON_WS, B("\n\t\t")}, {JSON_RSB, B(`]`)},
			{JSON_EOF, B("")},
		},
	}, {
		name:  "an object",
		input: `{"num":-0.123,"text":"foo bar","list":[{"key1":true},{"key2":null}]}`,
		want: []JSONToken{
			{JSON_LCB, B(`{`)},
			{JSON_KEY, B(`"num"`)}, {JSON_CLN, B(`:`)}, {JSON_NUM, B(`-0.123`)}, {JSON_COM, B(`,`)},
			{JSON_KEY, B(`"text"`)}, {JSON_CLN, B(`:`)}, {JSON_STR, B(`"foo bar"`)}, {JSON_COM, B(`,`)},
			{JSON_KEY, B(`"list"`)}, {JSON_CLN, B(`:`)},
			{JSON_LSB, B(`[`)}, {JSON_LCB, B(`{`)},
			{JSON_KEY, B(`"key1"`)}, {JSON_CLN, B(`:`)}, {JSON_TRU, B(`true`)},
			{JSON_RCB, B(`}`)}, {JSON_COM, B(`,`)}, {JSON_LCB, B(`{`)},
			{JSON_KEY, B(`"key2"`)}, {JSON_CLN, B(`:`)}, {JSON_NUL, B(`null`)},
			{JSON_RCB, B(`}`)}, {JSON_RSB, B(`]`)},
			{JSON_RCB, B(`}`)},
			{JSON_EOF, B("")},
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
			0:  {JSON_LCB, B(`{`)},
			1:  {JSON_WS, B("\n\t\t\t")},
			2:  {JSON_KEY, B(`"num"`)},
			3:  {JSON_CLN, B(`:`)},
			4:  {JSON_WS, B(" ")},
			5:  {JSON_NUM, B(`-0.123`)},
			6:  {JSON_COM, B(`,`)},
			7:  {JSON_WS, B("\n\t\t\t")},
			8:  {JSON_KEY, B(`"text"`)},
			9:  {JSON_CLN, B(`:`)},
			10: {JSON_WS, B(" ")},
			11: {JSON_STR, B(`"foo bar"`)},
			12: {JSON_COM, B(`,`)},
			13: {JSON_WS, B("\n\t\t\t")},
			14: {JSON_KEY, B(`"list"`)},
			15: {JSON_CLN, B(`:`)},
			16: {JSON_WS, B(" ")},
			17: {JSON_LSB, B(`[`)},
			18: {JSON_WS, B("\n\t\t\t\t")},
			19: {JSON_LCB, B(`{`)},
			20: {JSON_WS, B(" ")},
			21: {JSON_KEY, B(`"key1"`)},
			22: {JSON_CLN, B(`:`)},
			23: {JSON_WS, B(" ")},
			24: {JSON_TRU, B(`true`)},
			25: {JSON_WS, B(" ")},
			26: {JSON_RCB, B(`}`)},
			27: {JSON_COM, B(`,`)},
			28: {JSON_WS, B("\n\n\t\t\t\t")},
			29: {JSON_LCB, B(`{`)},
			30: {JSON_WS, B("\n\t\t\t\t\t")},
			31: {JSON_KEY, B(`"key2"`)},
			32: {JSON_CLN, B(`:`)},
			33: {JSON_WS, B(" ")},
			34: {JSON_NUL, B(`null`)},
			35: {JSON_WS, B("\n\t\t\t\t")},
			36: {JSON_RCB, B(`}`)},
			37: {JSON_WS, B("\n\t\t\t")},
			38: {JSON_RSB, B(`]`)},
			39: {JSON_WS, B("\n\t\t")},
			40: {JSON_RCB, B(`}`)},
			41: {JSON_EOF, B("")},
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
