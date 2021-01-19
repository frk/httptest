package httptest

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"reflect"

	"github.com/frk/compare"
	"github.com/frk/form"
)

// The Body type represents the contents of an HTTP request or response body.
type Body interface {
	// Value returns the underlying value of the Body interface.
	Value() interface{}
	// Reader returns an io.Reader that can be used to read the contents of the body.
	Reader() (io.Reader, error)
	// ContentType returns the media type (MIME) that describes the data contained in the body.
	ContentType() string
	// CompareContent returns the result of the comparison between the
	// Body's contents and the contents of the given io.Reader. The level
	// of strictness of the comparison depends on the implementation. If
	// the contents are equivalent the returned error will be nil, otherwise
	// the error will describe the negative result of the comparison.
	CompareContent(io.Reader) error
}

// The QueryEncoderBody is an interface that groups the QueryEncoder and Body interfaces.
type QueryEncoderBody interface {
	QueryEncoder
	Body
}

// JSON wraps the given value v and returns a Body that represents the value as
// json encoded data. The resulting Body uses encoding/json to encode and decode
// the given value, see the encoding/json documentation for more details.
func JSON(v interface{}) Body { return jsonbody{v} }

// XML wraps the given value v and returns a Body that represents the value as
// xml encoded data. The resulting Body uses encoding/xml to encode and decode
// the given value, see the encoding/xml documentation for more details.
func XML(v interface{}) Body { return xmlbody{v} }

// CSV wraps the given value v and returns a Body that represents the value as csv
// encoded data. The resulting Body uses encoding/csv to encode and decode the given
// value, see the encoding/csv documentation for more details.
func CSV(v [][]string) Body { return csvbody{v} }

// Form wraps the given value v and returns a Body that represents the value as form
// encoded data. At the moment the resulting Body uses github.com/frk/form to encode
// and decode the given value, see the package's documentation for more details.
func Form(v interface{}) QueryEncoderBody { return formbody{v} }

// Text wraps the given value v and returns a Body that represents the value as plain text.
func Text(v string) Body { return textbody{v} }

////////////////////////////////////////////////////////////////////////////////
// JSON Body
////////////////////////////////////////////////////////////////////////////////

// jsonbody implements the Body interface.
type jsonbody struct{ v interface{} }

const jsonContentType = "application/json"

// Value returns the underlying value of the jsonbody.
func (b jsonbody) Value() interface{} { return b.v }

// ContentType returns the media type (MIME) of the jsonbody which
// in this case will always be "application/json".
func (b jsonbody) ContentType() string { return jsonContentType }

// Reader returns an io.Reader that can be used to read the jsonbody's underlying
// value as json encoded data. Reader uses encoding/json's Marshal func to encode the
// underlying value, see the documentation on encoding/json's Marshal for more details.
func (b jsonbody) Reader() (io.Reader, error) {
	bs, err := json.Marshal(b.v)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(bs), nil
}

// CompareContent returns the result of the comparison between the jsonbody's
// underlying value and the given io.Reader. CompareContent uses encoding/json's
// Decoder.Decode to decode the Reader's contents into a newly allocated value
// of the same type as the jsonbody's underlying value, see the documentation
// on encoding/json's Decoder.Decode for more details.
//
// CompareContent does a "loose" comparison where it checks only whether the
// underlying value can be recreated from the given reader, it does not care
// about any additional data that the reader might contain.
func (b jsonbody) CompareContent(r io.Reader) error {
	rt := reflect.TypeOf(b.v)

	var isptr bool
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		isptr = true
	}

	v := reflect.New(rt).Interface()
	if err := json.NewDecoder(r).Decode(v); err != nil {
		return &testError{code: errResponseBody, err: err}
	}

	if !isptr {
		// if the underlying value is not a pointer get the
		// indirect of the reflect.Value of the decoded v.
		v = reflect.Indirect(reflect.ValueOf(v)).Interface()
	}

	cmp := compare.Config{ObserveFieldTag: "cmp"}
	if err := cmp.Compare(v, b.v); err != nil {
		return &testError{code: errResponseBody, err: err}
	}
	return nil
}

// for debugging
func (b jsonbody) String() string {
	bs, err := json.MarshalIndent(b.v, "", "  ")
	if err != nil {
		log.Println("frk/httptest:", err)
		return "[JSON ERROR]"
	}
	return string(bs)
}

////////////////////////////////////////////////////////////////////////////////
// XML Body
////////////////////////////////////////////////////////////////////////////////

// xmlbody implements the Body interface.
type xmlbody struct{ v interface{} }

// Value returns the underlying value of the xmlbody.
func (b xmlbody) Value() interface{} { return b.v }

const xmlContentType = "application/xml"

// ContentType returns the media type (MIME) of the xmlbody which
// in this case will always be "application/xml".
func (b xmlbody) ContentType() string { return xmlContentType }

// Reader returns an io.Reader that can be used to read the xmlbody's underlying
// value as xml encoded data. Reader uses encoding/xml's Marshal func to encode the
// underlying value, see the documentation on encoding/xml's Marshal for more details.
func (b xmlbody) Reader() (io.Reader, error) {
	bs, err := xml.Marshal(b.v)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(bs), nil
}

// CompareContent returns the result of the comparison between the xmlbody's
// underlying value and the given io.Reader. CompareContent uses encoding/xml's
// Decoder.Decode to decode the Reader's contents into a newly allocated value
// of the same type as the xmlbody's underlying value, see the documentation
// on encoding/xml's Decoder.Decode for more details.
//
// CompareContent does a "loose" comparison where it checks only whether the
// underlying value can be recreated from the given reader, it does not care
// about any additional data that the reader might contain.
func (b xmlbody) CompareContent(r io.Reader) error {
	rt := reflect.TypeOf(b.v)

	var isptr bool
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		isptr = true
	}

	v := reflect.New(rt).Interface()
	if err := xml.NewDecoder(r).Decode(v); err != nil {
		return &testError{code: errResponseBody, err: err}
	}

	if !isptr {
		// if the underlying value is not a pointer get the
		// indirect of the reflect.Value of the decoded v.
		v = reflect.Indirect(reflect.ValueOf(v)).Interface()
	}

	cmp := compare.Config{ObserveFieldTag: "cmp"}
	if err := cmp.Compare(v, b.v); err != nil {
		return &testError{code: errResponseBody, err: err}
	}
	return nil
}

// for debugging
func (b xmlbody) String() string {
	bs, err := xml.MarshalIndent(b.v, "", "  ")
	if err != nil {
		log.Println("frk/httptest:", err)
		return "[XML ERROR]"
	}
	return string(bs)
}

////////////////////////////////////////////////////////////////////////////////
// CSV Body
////////////////////////////////////////////////////////////////////////////////
const csvContentType = "text/csv"

// csvbody implements the Body interface.
type csvbody struct{ v [][]string }

// Value returns the underlying value of the csvbody.
func (b csvbody) Value() interface{} { return b.v }

// ContentType returns the media type (MIME) of the csvbody which
// in this case will always be text/csv.
func (b csvbody) ContentType() string { return csvContentType }

// Reader returns an io.Reader that can be used to read
// the csvbody's value as csv encoded data.
func (b csvbody) Reader() (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	if err := csv.NewWriter(buf).WriteAll(b.v); err != nil {
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), nil
}

// CompareContent returns the result of the comparison between
// the csvbody value and the given io.Reader. CompareContent uses
// encoding/csv's Decoder.Decode to decode the Reader's contents.
func (b csvbody) CompareContent(r io.Reader) error {
	rec, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return err
	}

	cmp := compare.Config{ObserveFieldTag: "cmp"}
	if err := cmp.Compare(rec, b.v); err != nil {
		return &testError{code: errResponseBody, err: err}
	}
	return nil
}

// for debugging..
func (b csvbody) String() string {
	buf := bytes.NewBuffer(nil)
	if err := csv.NewWriter(buf).WriteAll(b.v); err != nil {
		log.Println("frk/httptest:", err)
		return "[CSV ERROR]"
	}
	return buf.String()
}

////////////////////////////////////////////////////////////////////////////////
// Form Body
////////////////////////////////////////////////////////////////////////////////

// formbody implements the Body interface.
type formbody struct {
	v interface{}
}

// Value returns the underlying value of the formbody.
func (b formbody) Value() interface{} { return b.v }

const formContentType = "application/x-www-form-urlencoded"

// ContentType returns the media type (MIME) of the formbody which
// in this case will always be "application/x-www-form-urlencoded".
func (b formbody) ContentType() string { return formContentType }

// Reader returns an io.Reader that can be used to read the formbody's underlying
// value as form encoded data. Reader uses github.com/frk/form's Marshal func to
// encode the underlying value, see the documentation on that package's Marshal
// for more details.
func (b formbody) Reader() (io.Reader, error) {
	bs, err := form.Marshal(b.v)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(bs), nil
}

// CompareContent returns the result of the comparison between the formbody's
// underlying value and the given io.Reader. CompareContent uses github.com/frk/form's
// Decoder.Decode to decode the Reader's contents into a newly allocated value
// of the same type as the formbody's underlying value, see the documentation
// on github.com/frk/form's Decoder.Decode for more details.
//
// CompareContent does a "loose" comparison where it checks only whether the
// underlying value can be recreated from the given reader, it does not care
// about any additional data that the reader might contain.
func (b formbody) CompareContent(r io.Reader) error {
	rt := reflect.TypeOf(b.v)

	var isptr bool
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		isptr = true
	}

	v := reflect.New(rt).Interface()
	if err := form.NewDecoder(r).Decode(v); err != nil {
		return &testError{code: errResponseBody, err: err}
	}

	if !isptr {
		// if the underlying value is not a pointer get the
		// indirect of the reflect.Value of the decoded v.
		v = reflect.Indirect(reflect.ValueOf(v)).Interface()
	}

	cmp := compare.Config{ObserveFieldTag: "cmp"}
	if err := cmp.Compare(v, b.v); err != nil {
		return &testError{code: errResponseBody, err: err}
	}
	return nil
}

// QueryEncode implements the QueryEncoder interface.
func (b formbody) QueryEncode() string {
	bs, err := form.Marshal(b.v)
	if err != nil {
		log.Println("hit:", err)
		return "[FORM ERROR]"
	}
	return string(bs)
}

// for debugging...
func (b formbody) String() string {
	return b.QueryEncode()
}

////////////////////////////////////////////////////////////////////////////////
// Text Body
////////////////////////////////////////////////////////////////////////////////

// textbody implements the Body interface.
type textbody struct{ v string }

const textContentType = "text/plain"

// Value returns the underlying value of the textbody.
func (b textbody) Value() interface{} { return b.v }

// ContentType returns the media type (MIME) of the textbody which
// in this case will always be "text/plain".
func (b textbody) ContentType() string { return textContentType }

// Reader returns an io.Reader that can be used to read the textbody's underlying value.
func (b textbody) Reader() (io.Reader, error) {
	return bytes.NewReader([]byte(b.v)), nil
}

// CompareContent returns the result of the comparison between the textbody's
// underlying value and the given io.Reader.
func (b textbody) CompareContent(r io.Reader) error {
	v, err := ioutil.ReadAll(r)
	if err != nil {
		return &testError{code: errResponseBody, err: err}
	}
	if err := compare.Compare(string(v), b.v); err != nil {
		return &testError{code: errResponseBody, err: err}
	}
	return nil
}

// for debugging
func (b textbody) String() string {
	return b.v
}
