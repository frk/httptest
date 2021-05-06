package lexer

import (
	"testing"

	"github.com/frk/compare"
)

func TestCSV(t *testing.T) {
	// helper
	B := func(s string) []byte { return []byte(s) }

	tests := []struct {
		name  string
		input string
		want  []CSVToken
	}{{
		name:  "empty input",
		input: ``,
		want:  []CSVToken{{CSV_EOF, B("")}},
	}, {
		name:  "empty fields",
		input: `,,`,
		want: []CSVToken{
			{CSV_HED, B("")},
			{CSV_COM, B(",")},
			{CSV_HED, B("")},
			{CSV_COM, B(",")},
			{CSV_HED, B("")},
			{CSV_EOF, B("")},
		},
	}, {
		name:  "non-escaped fields",
		input: `foo,bar,baz`,
		want: []CSVToken{
			{CSV_HED, B("foo")},
			{CSV_COM, B(",")},
			{CSV_HED, B("bar")},
			{CSV_COM, B(",")},
			{CSV_HED, B("baz")},
			{CSV_EOF, B("")},
		},
	}, {
		name:  "escaped fields",
		input: "foo,\"b,a\"\"r\n\",baz",
		want: []CSVToken{
			{CSV_HED, B("foo")},
			{CSV_COM, B(",")},
			{CSV_HED, B("\"b,a\"\"r\n\"")},
			{CSV_COM, B(",")},
			{CSV_HED, B("baz")},
			{CSV_EOF, B("")},
		},
	}, {
		name:  "multiple records",
		input: `foo,bar` + "\n" + `baz,"qu,ux"`,
		want: []CSVToken{
			{CSV_HED, B("foo")},
			{CSV_COM, B(",")},
			{CSV_HED, B("bar")},
			{CSV_NL, B("\n")},
			{CSV_FLD, B("baz")},
			{CSV_COM, B(",")},
			{CSV_FLD, B(`"qu,ux"`)},
			{CSV_EOF, B("")},
		},
	}, {
		name:  "multiple records (empty fields)",
		input: `,bar,` + "\n" + `"",quux,`,
		want: []CSVToken{
			{CSV_HED, B("")},
			{CSV_COM, B(",")},
			{CSV_HED, B("bar")},
			{CSV_COM, B(",")},
			{CSV_HED, B("")},
			{CSV_NL, B("\n")},
			{CSV_FLD, B(`""`)},
			{CSV_COM, B(",")},
			{CSV_FLD, B("quux")},
			{CSV_COM, B(",")},
			{CSV_FLD, B("")},
			{CSV_EOF, B("")},
		},
	}}

	var agg_tokens = func(input string) (tokens []CSVToken) {
		ch := CSV([]byte(input))
		for {
			tok := <-ch
			tokens = append(tokens, tok)
			if tok.Type == CSV_EOF {
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
