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
	F1 string  `json:"fooBar" set:"required"`
	F2 float64 `json:"foo_bar" set:"optional"`
	F3 int     `set:"conditional"`
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
