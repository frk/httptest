package godoc

type (
	TestPara0 struct{}
	//
	TestPara1 struct{}
	// line 1
	// line 2
	TestPara2 struct{}
	// para 1
	//
	// para 2
	//
	//
	//
	// para 3
	TestPara3 struct{}
	/* para 1 */
	/* para 2 */
	/* para 3 */
	TestPara4 struct{}
	/*comment
	 */
	TestPara5 struct{}
	/*
		comment
	*/
	TestPara6 struct{}
	// comment
	/*
		block
		comment
	*/
	/* another block */
	TestPara7 struct{}
	// comment line
	//	indented line
	TestPara8 struct{}
	/*
		comment line
			indented line
	*/
	TestPara9 struct{}
)

type (
	// ``
	TestCode0 struct{}
	// `hello world`
	TestCode1 struct{}
	// `<code>hello world</code>`
	TestCode2 struct{}
	// `hello
	// world`
	TestCode3 struct{}
	// `hello
	//
	// world`
	TestCode4 struct{}
	// `*hello *world`
	TestCode5 struct{}
)

type (
	// *hello world*
	TestEm1 struct{}
	// *hello
	// world*
	TestEm2 struct{}
	// *hello
	//
	// world*
	TestEm3 struct{}
	// *hello <world>*
	TestEm4 struct{}
	// **hello world**
	TestEm5 struct{}
	// **hello world*
	TestEm6 struct{}
	// **hello world
	TestEm7 struct{}
	// hello **world**
	TestEm8 struct{}
)

type (
	// http://hello.world
	TestAnchor1 struct{}
	// click here: https://www.example.com
	TestAnchor2 struct{}
	// click [here](https://www.example.com)
	TestAnchor3 struct{}
)
