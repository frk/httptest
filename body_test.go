package httptest

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/frk/compare"
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

func Test_Body_ContentType(t *testing.T) {
	tests := []struct {
		body Body
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
			got := tt.body.ContentType()
			if got != tt.want {
				t.Errorf("got=%q; want=%q", got, tt.want)
			}
		})
	}
}

func Test_Body_Value(t *testing.T) {
	tests := []struct {
		body Body
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
			got := tt.body.Value()
			if e := compare.Compare(got, tt.want); e != nil {
				t.Error(e)
			}
		})
	}
}

func Test_Body_Reader(t *testing.T) {
	tests := []struct {
		body Body
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
		body Body
		data string
		want error
	}{
		////////////////////////////////////////////////////////////////
		// JSON
		////////////////////////////////////////////////////////////////
		{
			// same data as body, OK
			body: JSON(jsonstruct{"foo", 84, &jsonstruct{"bar", 22, nil}}),
			data: `{"A":"foo","B":84,"C":{"A":"bar","B":22,"C":null}}`,
			want: nil,
		}, {
			// additional fields that aren't expected by the body are OK
			body: JSON(jsonstruct{"foo", 84, &jsonstruct{"bar", 22, nil}}),
			data: `{"XYZ":[1,2,3],"A":"foo","B":84,"C":{"A":"bar","B":22,"C":null}}`,
			want: nil,
		}, {
			// slices are OK
			body: JSON([]jsonstruct{{"foo", 84, nil}, {"bar", 22, nil}}),
			data: `[{"A":"foo","B":84,"C":null},{"A":"bar","B":22,"C":null}]`,
			want: nil,
		}, {
			// pointers to slices are OK
			body: JSON(&[]*jsonstruct{{"foo", 84, nil}, {"bar", 22, nil}}),
			data: `[{"A":"foo","B":84,"C":null},{"A":"bar","B":22,"C":null}]`,
			want: nil,
		}, {
			// mismatched values NOT OK
			body: JSON(jsonstruct{A: "foo"}),
			data: `{"A":"bar"}`,
			want: &testError{code: errResponseBody, err: errors.New("<dummy>")},
		}, {
			// extra comma NOT OK
			body: JSON(jsonstruct{A: "foo"}),
			data: `{"A":"foo",}`,
			want: &testError{code: errResponseBody, err: errors.New("<dummy>")},
		}, {
			// data is not json NOT OK
			body: JSON(jsonstruct{A: "foo"}),
			data: `<A>foo</A>`,
			want: &testError{code: errResponseBody, err: errors.New("<dummy>")},
		},

		////////////////////////////////////////////////////////////////
		// XML
		////////////////////////////////////////////////////////////////
		{
			// same data as body, OK
			body: XML(xmlroot{[]xmlelem{{"foo", 84, &xmlelem{"bar", 22, nil}}}}),
			data: `<xmlroot><elem><A>foo</A><B>84</B><C><A>bar</A><B>22</B></C></elem></xmlroot>`,
			want: nil,
		}, {
			// additional fields that aren't expected by the body are OK
			body: XML(xmlroot{[]xmlelem{{"foo", 84, &xmlelem{"bar", 22, nil}}}}),
			data: `<xmlroot><elem><XYZ>123</XYZ><A>foo</A><B>84</B><C><A>bar</A><B>22</B></C></elem></xmlroot>`,
			want: nil,
		}, {
			// multi-value slices
			body: XML(xmlroot{[]xmlelem{{"foo", 84, nil}, {"bar", 22, nil}}}),
			data: `<xmlroot><elem><A>foo</A><B>84</B></elem><elem><A>bar</A><B>22</B></elem></xmlroot>`,
			want: nil,
		}, {
			// pointers are OK
			body: XML(&xmlroot{[]xmlelem{{"foo", 84, nil}, {"bar", 22, nil}}}),
			data: `<xmlroot><elem><A>foo</A><B>84</B></elem><elem><A>bar</A><B>22</B></elem></xmlroot>`,
			want: nil,
		}, {
			// mismatched values NOT OK
			body: XML(xmlelem{A: "foo"}),
			data: `<xmlelem><A>bar</A></xmlelem>`,
			want: &testError{code: errResponseBody, err: errors.New("<dummy>")},
		}, {
			// mismatched closing element NOT OK
			body: XML(xmlelem{A: "foo"}),
			data: `<xmlelem><A>bar</B></xmlelem>`,
			want: &testError{code: errResponseBody, err: errors.New("<dummy>")},
		}, {
			// missing closing element NOT OK
			body: XML(xmlelem{A: "foo"}),
			data: `<xmlelem><A>bar</A>`,
			want: &testError{code: errResponseBody, err: errors.New("<dummy>")},
		},

		////////////////////////////////////////////////////////////////
		// CSV
		////////////////////////////////////////////////////////////////
		{
			// same data as body, OK
			body: CSV([][]string{{"foo", "bar"}, {"123", "456"}}),
			data: "foo,bar\n123,456",
			want: nil,
		}, {
			body: CSV([][]string{
				{"a", "b", "c", "d"},
				{"foo, bar", "baz", "2 qux", ""},
				{"世界", "123", "0.002304", "foo \"bar\""},
			}),
			data: "a,b,c,d\n\"foo, bar\",baz,2 qux,\n世界,123,0.002304,\"foo \"\"bar\"\"\"",
			want: nil,
		}, {
			// mismatched values NOT OK
			body: CSV([][]string{{"foo", "bar"}, {"123", "456"}}),
			data: "foo,bar\n456,123",
			want: &testError{code: errResponseBody, err: errors.New("<dummy>")},
		},

		////////////////////////////////////////////////////////////////
		// Form
		////////////////////////////////////////////////////////////////
		{
			// same data as body
			body: Form(formstruct{"foo", 84, []float32{3.14, 42.0003}}),
			data: `A=foo&B=84&C=3.14&C=42.0003`,
			want: nil,
		}, {
			// additional fields that aren't expected by the body are ok
			body: Form(formstruct{"foo", 84, []float32{3.14, 42.0003}}),
			data: `A=foo&B=84&C=3.14&C=42.0003&D=true&E=foo+bar`,
			want: nil,
		}, {
			// pointer is ok
			body: Form(&formstruct{"foo", 84, []float32{3.14, 42.0003}}),
			data: `A=foo&B=84&C=3.14&C=42.0003&D=true&E=foo+bar`,
			want: nil,
		}, {
			// mismatched values
			body: Form(formstruct{A: "foo"}),
			data: `A=bar`,
			want: &testError{code: errResponseBody, err: errors.New("<dummy>")},
		}, {
			// not urlencoded form
			body: Form(formstruct{A: "foo"}),
			data: `{"A":"foo"}`,
			want: &testError{code: errResponseBody, err: errors.New("<dummy>")},
		},

		////////////////////////////////////////////////////////////////
		// Text
		////////////////////////////////////////////////////////////////
		{
			// same data as body
			body: Text("Hello, 世界"),
			data: `Hello, 世界`,
			want: nil,
		}, {
			// same data as body #2
			body: Text(""),
			data: ``,
			want: nil,
		}, {
			// mismatched values
			body: Text("Hello, 世界"),
			data: `Hello, World`,
			want: &testError{code: errResponseBody, err: errors.New("<dummy>")},
		}, {
			// mismatched values
			body: Text("Hello, 世界"),
			data: ``,
			want: &testError{code: errResponseBody, err: errors.New("<dummy>")},
		}, {
			// mismatched values
			body: Text(""),
			data: `Hello, 世界`,
			want: &testError{code: errResponseBody, err: errors.New("<dummy>")},
		},
	}

	cmp := compare.Config{ObserveFieldTag: "cmp"}

	for _, tt := range tests {
		name := fmt.Sprintf("%T", tt.body)
		t.Run(name, func(t *testing.T) {
			r := strings.NewReader(tt.data)
			err := tt.body.CompareContent(r)
			if e := cmp.Compare(err, tt.want); e != nil {
				t.Error(e)
			}
		})
	}
}
