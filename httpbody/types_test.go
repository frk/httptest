package httpbody

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/frk/compare"
	"github.com/frk/httptest"
	"github.com/frk/httptest/httpdoc"
)

func readallString(r io.Reader) string {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

type jsonstruct struct {
	A string
	B int
	C *jsonstruct
}

type xmlelem struct {
	A string
	B int
	C *xmlelem
}

type xmlroot struct {
	Elems []xmlelem `xml:"elem"`
}

type formstruct struct {
	A string
	B int
	C []float32
}

func Test_Body_Type(t *testing.T) {
	tests := []struct {
		body httptest.Body
		want string
	}{
		{JSON(nil), jsonContentType},
		{XML(nil), xmlContentType},
		{CSV(nil), csvContentType},
		{Form(nil), formContentType},
		{Text(""), textContentType},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%T", tt.body)
		t.Run(name, func(t *testing.T) {
			got := tt.body.Type()
			if got != tt.want {
				t.Errorf("got=%q; want=%q", got, tt.want)
			}
		})
	}
}

func Test_Body_Value(t *testing.T) {
	tests := []struct {
		body httptest.Body
		want interface{}
	}{{
		body: JSON(jsonstruct{"foo", 84, &jsonstruct{"bar", 22, nil}}),
		want: jsonstruct{"foo", 84, &jsonstruct{"bar", 22, nil}},
	}, {
		body: XML(xmlroot{[]xmlelem{{"foo", 84, &xmlelem{"bar", 22, nil}}}}),
		want: xmlroot{[]xmlelem{{"foo", 84, &xmlelem{"bar", 22, nil}}}},
	}, {
		body: CSV([][]string{{"foo", "bar"}, {"123", "456"}}),
		want: [][]string{{"foo", "bar"}, {"123", "456"}},
	}, {
		body: Form(formstruct{"foo", 84, []float32{3.14, 42.0003}}),
		want: formstruct{"foo", 84, []float32{3.14, 42.0003}},
	}, {
		body: Text("Hello, 世界"),
		want: "Hello, 世界",
	}}

	for _, tt := range tests {
		name := fmt.Sprintf("%T", tt.body)
		t.Run(name, func(t *testing.T) {
			got, err := tt.body.(httpdoc.Valuer).Value()
			if err != nil {
				t.Error(err)
			} else {
				if e := compare.Compare(got, tt.want); e != nil {
					t.Error(e)
				}
			}
		})
	}
}

func Test_Body_Reader(t *testing.T) {
	tests := []struct {
		body httptest.Body
		want string
	}{
		////////////////////////////////////////////////////////////////
		// JSON
		////////////////////////////////////////////////////////////////
		{
			body: JSON(jsonstruct{"foo", 84, &jsonstruct{"bar", 22, nil}}),
			want: `{"A":"foo","B":84,"C":{"A":"bar","B":22,"C":null}}`,
		}, {
			body: JSON([]jsonstruct{{"foo", 84, nil}, {"bar", 22, nil}}),
			want: `[{"A":"foo","B":84,"C":null},{"A":"bar","B":22,"C":null}]`,
		}, {
			body: JSON(jsonstruct{}),
			want: `{"A":"","B":0,"C":null}`,
		}, {
			body: JSON(nil),
			want: `null`,
		},

		////////////////////////////////////////////////////////////////
		// XML
		////////////////////////////////////////////////////////////////
		{
			body: XML(xmlroot{[]xmlelem{{"foo", 84, &xmlelem{"bar", 22, nil}}}}),
			want: `<xmlroot><elem><A>foo</A><B>84</B><C><A>bar</A><B>22</B></C></elem></xmlroot>`,
		}, {
			body: XML(xmlroot{[]xmlelem{{"foo", 84, nil}, {"bar", 22, nil}}}),
			want: `<xmlroot><elem><A>foo</A><B>84</B></elem><elem><A>bar</A><B>22</B></elem></xmlroot>`,
		}, {
			body: XML(xmlroot{}),
			want: `<xmlroot></xmlroot>`,
		}, {
			body: XML(nil),
			want: ``,
		},

		////////////////////////////////////////////////////////////////
		// CSV
		////////////////////////////////////////////////////////////////
		{
			body: CSV([][]string{{"foo", "bar"}, {"123", "456"}}),
			want: "foo,bar\n123,456\n",
		}, {
			body: CSV([][]string{
				{"a", "b", "c", "d"},
				{"foo, bar", "baz", "2 qux", ""},
				{"世界", "123", "0.002304", "foo \"bar\""},
			}),
			want: "a,b,c,d\n\"foo, bar\",baz,2 qux,\n世界,123,0.002304,\"foo \"\"bar\"\"\"\n",
		}, {
			body: CSV(nil),
			want: "",
		},

		////////////////////////////////////////////////////////////////
		// Form
		////////////////////////////////////////////////////////////////
		{
			body: Form(formstruct{"foo", 84, []float32{3.14, 42.0003}}),
			want: `A=foo&B=84&C=3.14&C=42.0003`,
		}, {
			body: Form(formstruct{}),
			want: `A=&B=0`,
		}, {
			body: Form(nil),
			want: ``,
		},

		////////////////////////////////////////////////////////////////
		// Text
		////////////////////////////////////////////////////////////////
		{
			body: Text("Hello, 世界"),
			want: "Hello, 世界",
		}, {
			body: Text(""),
			want: "",
		},
	}

	for i, tt := range tests {
		name := fmt.Sprintf("%T", tt.body)
		t.Run(name, func(t *testing.T) {
			r, err := tt.body.Reader()
			if err != nil {
				t.Error(err)
				return
			}

			got := readallString(r)
			if got != tt.want {
				t.Errorf("#%d: err got %q, want %q", i, got, tt.want)
			}
		})
	}
}

func Test_Body_CompareContent(t *testing.T) {
	tests := []struct {
		body    httptest.Body
		data    string
		wantErr bool
	}{
		////////////////////////////////////////////////////////////////
		// JSON
		////////////////////////////////////////////////////////////////
		{
			// same data as body, OK
			body:    JSON(jsonstruct{"foo", 84, &jsonstruct{"bar", 22, nil}}),
			data:    `{"A":"foo","B":84,"C":{"A":"bar","B":22,"C":null}}`,
			wantErr: false,
		}, {
			// additional fields that aren't expected by the body are OK
			body:    JSON(jsonstruct{"foo", 84, &jsonstruct{"bar", 22, nil}}),
			data:    `{"XYZ":[1,2,3],"A":"foo","B":84,"C":{"A":"bar","B":22,"C":null}}`,
			wantErr: false,
		}, {
			// slices are OK
			body:    JSON([]jsonstruct{{"foo", 84, nil}, {"bar", 22, nil}}),
			data:    `[{"A":"foo","B":84,"C":null},{"A":"bar","B":22,"C":null}]`,
			wantErr: false,
		}, {
			// pointers to slices are OK
			body:    JSON(&[]*jsonstruct{{"foo", 84, nil}, {"bar", 22, nil}}),
			data:    `[{"A":"foo","B":84,"C":null},{"A":"bar","B":22,"C":null}]`,
			wantErr: false,
		}, {
			// mismatched values NOT OK
			body:    JSON(jsonstruct{A: "foo"}),
			data:    `{"A":"bar"}`,
			wantErr: true,
		}, {
			// extra comma NOT OK
			body:    JSON(jsonstruct{A: "foo"}),
			data:    `{"A":"foo",}`,
			wantErr: true,
		}, {
			// data is not json NOT OK
			body:    JSON(jsonstruct{A: "foo"}),
			data:    `<A>foo</A>`,
			wantErr: true,
		},

		////////////////////////////////////////////////////////////////
		// XML
		////////////////////////////////////////////////////////////////
		{
			// same data as body, OK
			body:    XML(xmlroot{[]xmlelem{{"foo", 84, &xmlelem{"bar", 22, nil}}}}),
			data:    `<xmlroot><elem><A>foo</A><B>84</B><C><A>bar</A><B>22</B></C></elem></xmlroot>`,
			wantErr: false,
		}, {
			// additional fields that aren't expected by the body are OK
			body:    XML(xmlroot{[]xmlelem{{"foo", 84, &xmlelem{"bar", 22, nil}}}}),
			data:    `<xmlroot><elem><XYZ>123</XYZ><A>foo</A><B>84</B><C><A>bar</A><B>22</B></C></elem></xmlroot>`,
			wantErr: false,
		}, {
			// multi-value slices
			body:    XML(xmlroot{[]xmlelem{{"foo", 84, nil}, {"bar", 22, nil}}}),
			data:    `<xmlroot><elem><A>foo</A><B>84</B></elem><elem><A>bar</A><B>22</B></elem></xmlroot>`,
			wantErr: false,
		}, {
			// pointers are OK
			body:    XML(&xmlroot{[]xmlelem{{"foo", 84, nil}, {"bar", 22, nil}}}),
			data:    `<xmlroot><elem><A>foo</A><B>84</B></elem><elem><A>bar</A><B>22</B></elem></xmlroot>`,
			wantErr: false,
		}, {
			// mismatched values NOT OK
			body:    XML(xmlelem{A: "foo"}),
			data:    `<xmlelem><A>bar</A></xmlelem>`,
			wantErr: true,
		}, {
			// mismatched closing element NOT OK
			body:    XML(xmlelem{A: "foo"}),
			data:    `<xmlelem><A>bar</B></xmlelem>`,
			wantErr: true,
		}, {
			// missing closing element NOT OK
			body:    XML(xmlelem{A: "foo"}),
			data:    `<xmlelem><A>bar</A>`,
			wantErr: true,
		},

		////////////////////////////////////////////////////////////////
		// CSV
		////////////////////////////////////////////////////////////////
		{
			// same data as body, OK
			body:    CSV([][]string{{"foo", "bar"}, {"123", "456"}}),
			data:    "foo,bar\n123,456",
			wantErr: false,
		}, {
			body: CSV([][]string{
				{"a", "b", "c", "d"},
				{"foo, bar", "baz", "2 qux", ""},
				{"世界", "123", "0.002304", "foo \"bar\""},
			}),
			data:    "a,b,c,d\n\"foo, bar\",baz,2 qux,\n世界,123,0.002304,\"foo \"\"bar\"\"\"",
			wantErr: false,
		}, {
			// mismatched values NOT OK
			body:    CSV([][]string{{"foo", "bar"}, {"123", "456"}}),
			data:    "foo,bar\n456,123",
			wantErr: true,
		},

		////////////////////////////////////////////////////////////////
		// Form
		////////////////////////////////////////////////////////////////
		{
			// same data as body
			body:    Form(formstruct{"foo", 84, []float32{3.14, 42.0003}}),
			data:    `A=foo&B=84&C=3.14&C=42.0003`,
			wantErr: false,
		}, {
			// additional fields that aren't expected by the body are ok
			body:    Form(formstruct{"foo", 84, []float32{3.14, 42.0003}}),
			data:    `A=foo&B=84&C=3.14&C=42.0003&D=true&E=foo+bar`,
			wantErr: false,
		}, {
			// pointer is ok
			body:    Form(&formstruct{"foo", 84, []float32{3.14, 42.0003}}),
			data:    `A=foo&B=84&C=3.14&C=42.0003&D=true&E=foo+bar`,
			wantErr: false,
		}, {
			// mismatched values
			body:    Form(formstruct{A: "foo"}),
			data:    `A=bar`,
			wantErr: true,
		}, {
			// not urlencoded form
			body:    Form(formstruct{A: "foo"}),
			data:    `{"A":"foo"}`,
			wantErr: true,
		},

		////////////////////////////////////////////////////////////////
		// Text
		////////////////////////////////////////////////////////////////
		{
			// same data as body
			body:    Text("Hello, 世界"),
			data:    `Hello, 世界`,
			wantErr: false,
		}, {
			// same data as body #2
			body:    Text(""),
			data:    ``,
			wantErr: false,
		}, {
			// mismatched values
			body:    Text("Hello, 世界"),
			data:    `Hello, World`,
			wantErr: true,
		}, {
			// mismatched values
			body:    Text("Hello, 世界"),
			data:    ``,
			wantErr: true,
		}, {
			// mismatched values
			body:    Text(""),
			data:    `Hello, 世界`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%T", tt.body)
		t.Run(name, func(t *testing.T) {
			r := strings.NewReader(tt.data)
			err := tt.body.Compare(r)
			if tt.wantErr != (err != nil) {
				t.Errorf("wantErr=%t got=%v", tt.wantErr, err)
			}
		})
	}
}
