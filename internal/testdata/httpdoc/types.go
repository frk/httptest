package httpdoc

type T1 struct {
	F1 string
	F2 string
	F3 float64
	F4 float64
}

// foo bar
type T2a struct {
	// foo bar
	Foo string
	// bar baz
	Bar float64
}

type T2 struct {
	// Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod
	// tempor incididunt ut labore et dolore magna aliqua.
	//
	// Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi
	// ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit
	// in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
	//
	// Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia
	// deserunt mollit anim id est laborum.
	F1 string
	F2 float64 // this is just a line comment
}
