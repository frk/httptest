package comment

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/frk/httptest/internal/testdata/comment"
	"github.com/frk/httptest/internal/types"
)

var src *types.Source

func init() {
	var err error

	_, f, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filepath.Dir(f))
	if src, err = types.Load(dir); err != nil {
		panic(err)
	}
}

func TestToHTML(t *testing.T) {
	tests := []struct {
		comment []string
		want    string
	}{{
		comment: []string{`// foo bar`},
		want:    "<p>foo bar</p>\n",
	}, {
		comment: []string{"// foo", "// bar", "// baz"},
		want:    "<p>foo\nbar\nbaz</p>\n",
	}, {
		comment: []string{"// foo", "//\tbar", "//\tbaz"},
		want:    "<p>foo</p>\n<pre><code>bar\nbaz</code></pre>\n",
	}, {
		comment: []string{`/* foo
		bar
		baz */`}, //`
		want: "<p>foo\nbar\nbaz</p>\n",
	}, {
		comment: []string{`/*
		foo
		bar
		baz
		*/`}, //`
		want: "<p>foo\nbar\nbaz</p>\n",
	}, {
		comment: []string{`/* foo
			bar
		baz */`}, //`
		want: "<p>foo</p>\n<pre><code>bar</code></pre>\n<p>baz</p>\n",
	}, {
		comment: []string{`/*
		foo
			bar
		baz
		*/`}, //`
		want: "<p>foo</p>\n<pre><code>bar</code></pre>\n<p>baz</p>\n",
	}, {
		comment: []string{"foo **bar *baz http://www.xyz.com* click** quux"},
		want:    "<p>foo <strong>bar <em>baz <a href=\"http://www.xyz.com\">http://www.xyz.com</a></em> click</strong> quux</p>\n",
	}, {
		comment: []string{"foo **bar *baz http://www.xyz.com** click* quux"},
		want:    "<p>foo <strong>bar *baz <a href=\"http://www.xyz.com\">http://www.xyz.com</a></strong> click* quux</p>\n",
	}}

	for _, tt := range tests {
		got, err := ToHTML(tt.comment)
		if err != nil {
			t.Error(err)
		} else if got != tt.want {
			t.Errorf("text: %q\n got: %q\nwant: %q", tt.comment, got, tt.want)
		}
	}
}

type typeDescTests []struct {
	v    interface{}
	want string
	skip bool
}

func (tests typeDescTests) Run(t *testing.T) {
	for i, tt := range tests {
		if tt.skip {
			continue
		}
		typ := src.TypeOf(tt.v)
		if false {
			fmt.Println(typ)
		}
		got, err := ToHTML(typ.Doc)
		if err != nil {
			t.Errorf("%d: error got:%v, want:%v", i, err, nil)
		} else if got != tt.want {
			t.Errorf("%d\ngot:%q,\nwnt:%q", i, got, tt.want)
		}
	}
}

func TestTypeDesc_para(t *testing.T) {
	tests := typeDescTests{{
		v:    comment.TestPara0{},
		want: "",
	}, {
		v:    comment.TestPara1{},
		want: "",
	}, {
		v:    comment.TestPara2{},
		want: "<p>line 1\nline 2</p>\n",
	}, {
		v:    comment.TestPara3{},
		want: "<p>para 1</p>\n<p>para 2</p>\n<p>para 3</p>\n",
	}, {
		v:    comment.TestPara4{},
		want: "<p>para 1\npara 2\npara 3</p>\n",
	}, {
		v:    comment.TestPara5{},
		want: "<p>comment</p>\n",
	}, {
		v:    comment.TestPara6{},
		want: "<p>comment</p>\n",
	}, {
		v:    comment.TestPara7{},
		want: "<p>comment</p>\n<p>block\ncomment</p>\n<p>another block</p>\n",
	}, {
		v:    comment.TestPara8{},
		want: "<p>comment line</p>\n<pre><code>indented line</code></pre>\n",
	}, {
		v:    comment.TestPara9{},
		want: "<p>comment line</p>\n<pre><code>indented line</code></pre>\n",
	}}

	tests.Run(t)
}

func TestTypeDesc_code(t *testing.T) {
	tests := typeDescTests{{
		v:    comment.TestCode0{},
		want: "<p><code></code></p>\n",
	}, {
		v:    comment.TestCode1{},
		want: "<p><code>hello world</code></p>\n",
	}, {
		v:    comment.TestCode2{},
		want: "<p><code>&lt;code&gt;hello world&lt;/code&gt;</code></p>\n",
	}, {
		v:    comment.TestCode3{},
		want: "<p><code>hello\nworld</code></p>\n",
	}, {
		v:    comment.TestCode4{},
		want: "<p>`hello</p>\n<p>world`</p>\n",
	}, {
		v:    comment.TestCode5{},
		want: "<p><code>*hello *world</code></p>\n",
	}}

	tests.Run(t)
}

func TestTypeDesc_em(t *testing.T) {
	tests := typeDescTests{{
		v:    comment.TestEm1{},
		want: "<p><em>hello world</em></p>\n",
	}, {
		v:    comment.TestEm2{},
		want: "<p><em>hello\nworld</em></p>\n",
	}, {
		v:    comment.TestEm3{},
		want: "<p>*hello</p>\n<p>world*</p>\n",
	}, {
		v:    comment.TestEm4{},
		want: "<p><em>hello <world></em></p>\n",
	}, {
		v:    comment.TestEm5{},
		want: "<p><strong>hello world</strong></p>\n",
	}, {
		v:    comment.TestEm6{},
		want: "<p>**hello world*</p>\n",
	}, {
		v:    comment.TestEm7{},
		want: "<p>**hello world</p>\n",
	}, {
		v:    comment.TestEm8{},
		want: "<p>hello <strong>world</strong></p>\n",
	}}

	tests.Run(t)
}

func TestTypeDesc_anchor(t *testing.T) {
	tests := typeDescTests{{
		v: comment.TestAnchor1{}, want: "<p><a href=\"http://hello.world\">http://hello.world</a></p>\n",
	}, {
		v: comment.TestAnchor2{}, want: "<p>click here: <a href=\"https://www.example.com\">https://www.example.com</a></p>\n",
	}, {
		v: comment.TestAnchor3{}, want: "<p>click <a href=\"https://www.example.com\">here</a></p>\n",
	}}

	tests.Run(t)
}

func TestLexer(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []item
	}{
		{"text only", "hello world", []item{
			{itemParaStart, "", 0},
			{itemText, "hello world", 0},
			{itemEOF, "", 11},
		}},
		{"with raw string", "hello `raw` world", []item{
			{itemParaStart, "", 0},
			{itemText, "hello ", 0},
			{itemRaw, "`raw`", 6},
			{itemText, " world", 11},
			{itemEOF, "", 17},
		}},
		{"with asterisk", "hello *star* world", []item{
			{itemParaStart, "", 0},
			{itemText, "hello ", 0},
			{itemStar, "*", 6},
			{itemText, "star", 7},
			{itemStar, "*", 11},
			{itemText, " world", 12},
			{itemEOF, "", 18},
		}},
		{"with double asterisk", "hello **double star** world", []item{
			{itemParaStart, "", 0},
			{itemText, "hello ", 0},
			{itemStar2, "**", 6},
			{itemText, "double star", 8},
			{itemStar2, "**", 19},
			{itemText, " world", 21},
			{itemEOF, "", 27},
		}},
		{"URL only", "http://example.com", []item{
			{itemParaStart, "", 0},
			{itemURL, "http://example.com", 0},
			{itemEOF, "", 18},
		}},
		{"Named URL", "[an example](http://example.com)", []item{
			{itemParaStart, "", 0},
			{itemNamedURL, "[an example](http://example.com)", 0},
			{itemEOF, "", 32},
		}},
		{"URL with text", "click here: http://example.com ok?", []item{
			{itemParaStart, "", 0},
			{itemText, "click here: ", 0},
			{itemURL, "http://example.com", 12},
			{itemText, " ok?", 30},
			{itemEOF, "", 34},
		}},
		{"named URL with text", "click [here](http://example.com) ok?", []item{
			{itemParaStart, "", 0},
			{itemText, "click ", 0},
			{itemNamedURL, "[here](http://example.com)", 6},
			{itemText, " ok?", 32},
			{itemEOF, "", 36},
		}},
		{"indented text", "hello\n\tindented\nworld", []item{
			{itemParaStart, "", 0},
			{itemText, "hello", 0},
			{itemIdent, "\tindented\n", 6},
			{itemParaStart, "", 16},
			{itemText, "world", 16},
			{itemEOF, "", 21},
		}},
		{"more indented text", "hello\n\t\tindented\n space-indented\nworld", []item{
			{itemParaStart, "", 0},
			{itemText, "hello", 0},
			{itemIdent, "\t\tindented\n space-indented\n", 6},
			{itemParaStart, "", 33},
			{itemText, "world", 33},
			{itemEOF, "", 38},
		}},
		{"multi paragraph text", "foo\n\nbar\n\nbaz\n\nquux", []item{
			{itemParaStart, "", 0},
			{itemText, "foo", 0},
			{itemParaStart, "", 5},
			{itemText, "bar", 5},
			{itemParaStart, "", 10},
			{itemText, "baz", 10},
			{itemParaStart, "", 15},
			{itemText, "quux", 15},
			{itemEOF, "", 19},
		}},
		{"multi line text", "foo\nbar\nbaz\nquux", []item{
			{itemParaStart, "", 0},
			{itemText, "foo", 0},
			{itemText, "\nbar", 3},
			{itemText, "\nbaz", 7},
			{itemText, "\nquux", 11},
			{itemEOF, "", 16},
		}},
		{"multi line text", "foo\nbar\n\nbaz\nquux", []item{
			{itemParaStart, "", 0},
			{itemText, "foo", 0},
			{itemText, "\nbar", 3},
			{itemParaStart, "", 9},
			{itemText, "baz", 9},
			{itemText, "\nquux", 12},
			{itemEOF, "", 17},
		}},
		{"stars url item", "foo *bar **baz http://www.xyz.com*** quux", []item{
			{itemParaStart, "", 0},
			{itemText, "foo ", 0},
			{itemStar, "*", 4},
			{itemText, "bar ", 5},
			{itemStar2, "**", 9},
			{itemText, "baz ", 11},
			{itemURL, "http://www.xyz.com", 15},
			{itemStar2, "**", 33},
			{itemStar, "*", 35},
			{itemText, " quux", 36},
			{itemEOF, "", 41},
		}},
		{"stars url item 2", "foo **bar *baz http://www.xyz.com** click* quux", []item{
			{itemParaStart, "", 0},
			{itemText, "foo ", 0},
			{itemStar2, "**", 4},
			{itemText, "bar ", 6},
			{itemStar, "*", 10},
			{itemText, "baz ", 11},
			{itemURL, "http://www.xyz.com", 15},
			{itemStar2, "**", 33},
			{itemText, " click", 35},
			{itemStar, "*", 41},
			{itemText, " quux", 42},
			{itemEOF, "", 47},
		}},
		{"stars named url item 2", "foo **bar *baz [xyz](http://www.xyz.com)** click* quux", []item{
			{itemParaStart, "", 0},
			{itemText, "foo ", 0},
			{itemStar2, "**", 4},
			{itemText, "bar ", 6},
			{itemStar, "*", 10},
			{itemText, "baz ", 11},
			{itemNamedURL, "[xyz](http://www.xyz.com)", 15},
			{itemStar2, "**", 40},
			{itemText, " click", 42},
			{itemStar, "*", 48},
			{itemText, " quux", 49},
			{itemEOF, "", 54},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collect(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\n got: %+v\nwant: %+v", got, tt.want)
			}
		})
	}
}

func collect(input string) (items []item) {
	lx := lex(input)
	for {
		it := <-lx.items
		items = append(items, it)
		if it.typ == itemEOF {
			break
		}
	}
	return
}

func TestParser(t *testing.T) {
	tests := []struct {
		text string
		want *node
	}{
		{"hello world", &node{typ: nodeRoot, children: []*node{
			{typ: nodePara, children: []*node{{typ: nodeText, data: "hello world"}}}}}},
		{"http://hello.world", &node{typ: nodeRoot, children: []*node{
			{typ: nodePara, children: []*node{{typ: nodeAnchor, data: "http://hello.world", href: "http://hello.world"}}}}}},
		{"foo *bar* baz", &node{typ: nodeRoot, children: []*node{
			{typ: nodePara, children: []*node{
				{typ: nodeText, pos: 0, data: "foo "},
				{typ: nodeEm, pos: 4, children: []*node{{typ: nodeText, pos: 5, data: "bar"}}},
				{typ: nodeText, pos: 9, data: " baz"}}}}}},
		{"foo **bar** baz", &node{typ: nodeRoot, children: []*node{
			{typ: nodePara, children: []*node{
				{typ: nodeText, pos: 0, data: "foo "},
				{typ: nodeStrong, pos: 4, children: []*node{{typ: nodeText, pos: 6, data: "bar"}}},
				{typ: nodeText, pos: 11, data: " baz"}}}}}},
		{"foo **", &node{typ: nodeRoot, children: []*node{
			{typ: nodePara, children: []*node{
				{typ: nodeText, pos: 0, data: "foo "},
				{typ: nodeText, pos: 4, data: "**"}}}}}},
		{"foo **bar baz", &node{typ: nodeRoot, children: []*node{
			{typ: nodePara, children: []*node{
				{typ: nodeText, pos: 0, data: "foo "},
				{typ: nodeText, pos: 4, data: "**"},
				{typ: nodeText, pos: 6, data: "bar baz"}}}}}},
		{"foo `bar` baz", &node{typ: nodeRoot, children: []*node{
			{typ: nodePara, children: []*node{
				{typ: nodeText, pos: 0, data: "foo "},
				{typ: nodeCode, pos: 4, data: "bar"},
				{typ: nodeText, pos: 9, data: " baz"}}}}}},
		{"foo http://bar.com baz", &node{typ: nodeRoot, children: []*node{
			{typ: nodePara, children: []*node{
				{typ: nodeText, pos: 0, data: "foo "},
				{typ: nodeAnchor, pos: 4, data: "http://bar.com", href: "http://bar.com"},
				{typ: nodeText, pos: 18, data: " baz"}}}}}},
		{"foo [bar](http://bar.com) baz", &node{typ: nodeRoot, children: []*node{
			{typ: nodePara, children: []*node{
				{typ: nodeText, pos: 0, data: "foo "},
				{typ: nodeAnchor, pos: 4, data: "bar", href: "http://bar.com"},
				{typ: nodeText, pos: 25, data: " baz"}}}}}},
		{"foo\n\tbar\nbaz", &node{typ: nodeRoot, children: []*node{
			{typ: nodePara, children: []*node{{typ: nodeText, pos: 0, data: "foo"}}},
			{typ: nodePre, pos: 4, data: "bar"},
			{typ: nodePara, pos: 9, children: []*node{{typ: nodeText, pos: 9, data: "baz"}}}}}},
		{"foo **bar *baz http://www.xyz.com** click* quux", &node{typ: nodeRoot, children: []*node{
			{typ: nodePara, children: []*node{
				{typ: nodeText, data: "foo "},
				{typ: nodeStrong, pos: 4, children: []*node{
					{typ: nodeText, pos: 6, data: "bar "},
					{typ: nodeText, pos: 10, data: "*"},
					{typ: nodeText, pos: 11, data: "baz "},
					{typ: nodeAnchor, pos: 15, data: "http://www.xyz.com", href: "http://www.xyz.com"},
				}},
				{typ: nodeText, pos: 35, data: " click"},
				{typ: nodeText, pos: 41, data: "*"},
				{typ: nodeText, pos: 42, data: " quux"},
			}},
		}}},
	}

	for _, tt := range tests {
		root := parsedoc(tt.text)
		if !reflect.DeepEqual(root, tt.want) {
			t.Errorf("\n got: %s\nwant: %s\n", root, tt.want)
		}
	}
}

func compareNodes(got, want *node) error {
	if got.typ != want.typ {
		return fmt.Errorf("got typ %d, want typ %d", got.typ, want.typ)
	}
	if got.data != want.data {
		return fmt.Errorf("got data %s, want data %s", got.data, want.data)
	}
	if gotc, wantc := len(got.children), len(want.children); gotc != wantc {
		return fmt.Errorf("got num of children %d, want num of children %d", gotc, wantc)
	}

	for i, c := range got.children {
		if err := compareNodes(c, want.children[i]); err != nil {
			return err
		}
	}
	return nil
}
