package markup

import (
	"testing"

	"github.com/frk/compare"
)

func TestCSV(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{{
		name: "empty input",
		data: ``,
		want: ``,
	}, {
		name: "empty string",
		data: `""`,
		want: `<span class="token csv-header">""</span>`,
	}, {
		name: "empty string",
		data: `foo,bar,baz`,
		want: `<span class="token csv-header">foo</span>` +
			`<span class="token csv-com">,</span>` +
			`<span class="token csv-header">bar</span>` +
			`<span class="token csv-com">,</span>` +
			`<span class="token csv-header">baz</span>`,
	}, {
		name: "empty string",
		data: `foo,bar` + "\n" + `baz,quux`,
		want: `<span class="token csv-header">foo</span>` +
			`<span class="token csv-com">,</span>` +
			`<span class="token csv-header">bar</span>` +
			"\n" +
			`<span class="token csv-field">baz</span>` +
			`<span class="token csv-com">,</span>` +
			`<span class="token csv-field">quux</span>`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CSV([]byte(tt.data))
			if e := compare.Compare(got, tt.want); e != nil {
				t.Error(e)
			}
		})
	}
}
