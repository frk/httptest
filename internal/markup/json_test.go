package markup

import (
	"testing"

	"github.com/frk/compare"
)

func TestJSON(t *testing.T) {
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
		want: `<span class="token json-str">&#34;&#34;</span>`,
	}, {
		name: "an array",
		data: `[
			1,
			12,
			123,
			"foo bar",
			null,
			{
				"key1": -0.009876,
				"key2": "abcdefg"
			},
			false
		]`,
		want: `<span class="token json-lsb">[</span>
			<span class="token json-num">1</span><span class="token json-com">,</span>
			<span class="token json-num">12</span><span class="token json-com">,</span>
			<span class="token json-num">123</span><span class="token json-com">,</span>
			<span class="token json-str">&#34;foo bar&#34;</span><span class="token json-com">,</span>
			<span class="token json-nul">null</span><span class="token json-com">,</span>
			<span class="token json-lcb">{</span>
				<span class="token json-key-q">"</span><span class="token json-key-t">key1</span><span class="token json-key-q">"</span><span class="token json-cln">:</span> <span class="token json-num">-0.009876</span><span class="token json-com">,</span>
				<span class="token json-key-q">"</span><span class="token json-key-t">key2</span><span class="token json-key-q">"</span><span class="token json-cln">:</span> <span class="token json-str">&#34;abcdefg&#34;</span>
			<span class="token json-rcb">}</span><span class="token json-com">,</span>
			<span class="token json-fal">false</span>
		<span class="token json-rsb">]</span>`,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JSON([]byte(tt.data))
			if e := compare.Compare(got, tt.want); e != nil {
				t.Error(e)
			}
		})
	}
}
