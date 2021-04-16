package httpdoc

type T1 struct {
	F1 string  `doc:"required"`
	F2 string  `doc:"optional"`
	F3 float64 `doc:",required"`
	F4 float64
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

type T3 struct {
	F1 string  `json:"fooBar" set:"required" validation:"len:8:128"`
	F2 float64 `json:"foo_bar" set:"optional" validation:"max:0.7"`
	F3 int     `set:"conditional" validation:"min:21"`
	F4 bool    `json:"foo-bar"`
}

type T4 struct {
	F1 struct {
		G1 bool
		G2 struct {
			H1 int
			H2 struct {
				I1 string
				I2 string
			}
			H3 bool
		}
	}
	F2 float32
}

type T5 struct {
	F1 struct {
		G1 bool `json:"g_1"`
		G2 struct {
			H1 int `json:"h_1"`
			H2 struct {
				I1 string `json:"i_1"`
				I2 string `json:"i_2"`
			} `json:"h_2"`
			H3 bool `json:"h_3"`
		} `json:"g_2"`
	} `json:"f_1"`
	F2 float32 `json:"f_2"`
}

type T6 struct {
	F1 struct {
		G1 bool `json:"g_1"`
		G2 struct {
			H1 int `json:"h_1"`
			H2 struct {
				I1 string
				I2 string `json:"i_2"`
			} `json:"h_2"`
			H3 bool
		}
	} `json:"f_1"`
	F2 float32
}

type T7 struct {
	F1 E1
}

type E1 string

const (
	// Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod
	// tempor incididunt ut labore et dolore magna aliqua.
	E1_foo E1 = "foo"
	// Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi
	// ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit
	// in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
	E1_bar E1 = "bar"
	E1_baz E1 = "baz" // this is just a line comment
)

// This is the V1 type docs.
type V1 struct {
	F1 string `json:"f1"`
	F2 int32  `json:"f2"`
}

type E2 int64

const (
	_ E2 = 1<<iota*2 - 3
	// Lorem ipsum dolor sit amet, consectetur adipiscing elit.
	E2_foo
	// Duis aute irure dolor in reprehenderit in voluptate velit esse cillum
	// dolore eu fugiat nulla pariatur.
	E2_bar1
	E2_bar2
	_
	_
	E2_baz E2 = 0xBadFace // this is just a line comment
)

const E2_quux E2 = 8 * (5 + 3) / (3 ^ (2 - 23)) * -1

type T8 struct {
	F1 E2
}

type H1RES struct {
	// The content type of the response.
	ContentType string `json:"Content-Type"`
	// The Cookie will contain the created session id.
	SetCookie string `json:"Set-Cookie"`
}

type H1REQ struct {
	// The Authorization header must contain the bearer token.
	Authorization string `json:"Authorization" doc:"required"`
}

type P1 struct {
	// The id of the Foo to which the Bar belongs.
	FooID int `json:"foo_id" doc:"required"`
	// The UUID of the Bar to retrieve.
	BarUUID string `json:"bar_uuid" doc:"required"`
}

type Q1 struct {
	// The search string.
	Q string `json:"q" doc:"optional"`
	// The max number of results to return, if not present it will default to 32.
	Limit int `json:"limit" doc:"optional"`
}

// This is an Auth type and its documentation that needs to be extracted
// and turned into HTML for rendering.
//
// This is a pragraph that doesn't say anything useful.
type A1 struct{}

type FT1 struct {
	// Lorem ipsum dolor sit amet, consectetur **adipiscing elit**, sed do eiusmod
	// tempor *incididunt* ut labore et dolore magna aliqua.
	//
	// Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi
	//	ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit
	//	in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
	//
	// Excepteur `sint occaecat cupidatat non proident`, sunt in culpa qui officia
	// deserunt mollit anim id est laborum. [test link](https://example.com).
	Foo struct {
		// This is just a bar.
		Bar struct {
			/*
				... burn, like fabulous yellow roman candles exploding
				like spiders across the stars ...
			*/
			Baz string `set:"required" validation:"len:10:128"`
		} `json:"b_a_r"`
	} `json:"FOO"`
}
