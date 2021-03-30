package types

type T1 struct{}

type T2 struct {
	F string
}

type T3 struct {
	F1 string
	F2 string
	F3 float64
	F4 float64
}

type T4 struct {
	F1 string
	F2 int
	F3 *T2
}

type T5a struct {
	F1 T2
	F2 T2
}

type T5b struct {
	F1 *T2
	F2 *T2
}

type T6 struct {
	F1 *T6
	F2 *T6
}

type T7 struct {
	F interface{}
}

type T8 struct {
	F1 interface{}
	F2 interface{}
}

type T9 struct {
	F1 *T9
	F2 interface{}
}

////////////////////////////////////////////////////////////////////////////////

// S1 decl doc
type S1 string

type S2 string // S2 comment

// S3,S4,S5 decl doc
type (
	S3 string

	// S4 spec doc
	S4 string

	S5 string // S5 comment (this one's ignored in favor of decl doc)
)

/*
	S6 decl doc line 1
	S6 decl doc line 2
*/
type S6 string

/*
	S7 decl doc line 1
	S7 decl doc line 2
*/
// S7 decl doc line 3
// S7 decl doc line 4
type S7 string

// S8 decl doc
type S8 struct {
	// S8.F1 doc line 1
	// S8.F1 doc line 2
	F1 string // S8.F1 comment (ignored in favor of doc lines)
	F2 string // S8.F2 comment
	F3 string /* S8.F3 comment line 1
	S8.F3 comment line 2
	S8.F3 comment line 3
	*/
}

////////////////////////////////////////////////////////////////////////////////

type C1 int

// C1 const decl doc
const (
	// c1a spec doc line 1
	// c1a spec doc line 2
	c1a C1 = 1<<iota*2 - 3
	c1b
	_
	c1c // c1c comment line
)

type C2 string

const (
	c2a C2 = "foo"
	c2b    = C2("bar")
	c2c    = "ignored because untyped"
)

const c2d = C2("baz") // c2d comment line
