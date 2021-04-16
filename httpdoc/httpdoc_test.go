package httpdoc

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/frk/compare"
	"github.com/frk/httptest"
	"github.com/frk/httptest/internal/page"
	"github.com/frk/httptest/internal/testdata/httpdoc"
	"github.com/frk/tagutil"
)

func TestBuild(t *testing.T) {
	_, f, _, _ := runtime.Caller(0)
	workdir := filepath.Dir(f)
	srclocal, err := findRootDir(workdir)
	if err != nil {
		t.Error(err)
		return
	}

	srcremote := "https://github.com/frk/httptest/tree/master/"
	srclink := SourceURLFunc(srclocal, srcremote)

	testdatadir := filepath.Dir(workdir) + "/internal/testdata"
	testFile, err := os.Open(testdatadir + "/test.html")
	if err != nil {
		t.Error(err)
		return
	}
	defer testFile.Close()

	tests := []struct {
		skip bool
		file string
		cfg  Config
		mode page.TestMode
		toc  []*ArticleGroup
	}{{
		///////////////////////////////////////////////////////////////
		// Page
		/////////////////////////////////////////////////////////////////
		file: "page_empty",
		toc:  []*ArticleGroup{},
	}, {
		file: "page_empty",
		toc:  []*ArticleGroup{{Name: "Group Name"}},
	}, {
		file: "page_with_article",
		toc: []*ArticleGroup{{
			Name:     "Group Name",
			Articles: []*Article{{Title: "Article Title"}},
		}},
	}, {
		file: "page_with_endpoint",
		toc: []*ArticleGroup{{
			Name: "Group Name",
			Articles: []*Article{{
				Title: "Article Title",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"POST", "/api/foos"},
				}},
			}},
		}},
	}, {
		///////////////////////////////////////////////////////////////
		// Sidebar
		/////////////////////////////////////////////////////////////////
		file: "sidebar_from_articles",
		mode: page.SidebarTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article 1",
				SubArticles: []*Article{{
					Title: "Sub Article 1",
				}, {
					Title: "Sub Article 2",
					SubArticles: []*Article{{
						Title: "Sub Sub Article 1",
					}},
				}},
			}, {
				Title: "Article 2",
			}},
		}, {
			Name: "Article Group 2",
			Articles: []*Article{{
				Title: "Article 3",
			}},
		}},
	}, {
		file: "sidebar_from_endpoints",
		mode: page.SidebarTest,
		toc: []*ArticleGroup{{
			Name: "Endpoint Group 1",
			Articles: []*Article{{
				Title: "Article 1",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"POST", "/api/foos"},
					Desc:     "Create a Foo",
				}, {
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Desc:     "List Foos",
				}, {
					Endpoint: httptest.Endpoint{"GET", "/api/foos/{id}"},
					Desc:     "Get a Foo",
				}, {
					Endpoint: httptest.Endpoint{"DELETE", "/api/foos"},
					Desc:     "Delete a Foo",
				}},
			}},
		}},
	}, {
		file: "sidebar_from_mix",
		mode: page.SidebarTest,
		toc: []*ArticleGroup{{
			Name: "Endpoint Group 1",
			Articles: []*Article{{
				Title: "Article 1",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"POST", "/api/foos"},
					Desc:     "Create a Foo",
				}, {
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Desc:     "List Foos",
				}},
				SubArticles: []*Article{{
					Title: "Sub Article 1",
				}, {
					Title: "Sub Article 2",
					SubArticles: []*Article{{
						Title: "Sub Sub Endpoints 1",
						TestGroups: []*httptest.TestGroup{{
							Endpoint: httptest.Endpoint{"POST", "/api/foos/{id}/bars"},
							Desc:     "Create a FooBar",
						}, {
							Endpoint: httptest.Endpoint{"GET", "/api/foos/{id}/bars"},
							Desc:     "List FooBars",
						}},
					}},
				}},
			}},
		}, {
			Name: "Article Group 2",
			Articles: []*Article{{
				Title: "Article 3",
			}},
		}},
	}, {
		///////////////////////////////////////////////////////////////
		// Content
		/////////////////////////////////////////////////////////////////
		file: "content_from_articles",
		mode: page.ContentTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article 1",
				SubArticles: []*Article{{
					Title: "Sub Article 1",
				}, {
					Title: "Sub Article 2",
					SubArticles: []*Article{{
						Title: "Sub Sub Article 1",
					}},
				}},
			}, {
				Title: "Article 2",
			}},
		}, {
			Name: "Article Group 2",
			Articles: []*Article{{
				Title: "Article 3",
			}},
		}},
	}, {
		file: "content_from_endpoints",
		mode: page.ContentTest,
		toc: []*ArticleGroup{{
			Name: "Endpoint Group 1",
			Articles: []*Article{{
				Title: "Article 1",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"POST", "/api/foos"},
					Desc:     "Create a Foo",
				}, {
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Desc:     "List Foos",
				}, {
					Endpoint: httptest.Endpoint{"GET", "/api/foos/{id}"},
					Desc:     "Get a Foo",
				}, {
					Endpoint: httptest.Endpoint{"DELETE", "/api/foos"},
					Desc:     "Delete a Foo",
				}},
			}},
		}},
	}, {
		file: "content_from_mix",
		mode: page.ContentTest,
		toc: []*ArticleGroup{{
			Name: "Endpoint Group 1",
			Articles: []*Article{{
				Title: "Article 1",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"POST", "/api/foos"},
					Desc:     "Create a Foo",
				}, {
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Desc:     "List Foos",
				}},
				SubArticles: []*Article{{
					Title: "Sub Article 1",
				}, {
					Title: "Sub Article 2",
					SubArticles: []*Article{{
						Title: "Sub Sub Endpoints 1",
						TestGroups: []*httptest.TestGroup{{
							Endpoint: httptest.Endpoint{"GET", "/api/foos/{id}/bars"},
							Desc:     "List FooBars",
						}},
					}},
				}},
			}},
		}, {
			Name: "Article Group 2",
			Articles: []*Article{{
				Title: "Article 3",
			}},
		}},
	}, {
		///////////////////////////////////////////////////////////////
		// Article Text Column
		/////////////////////////////////////////////////////////////////
		file: "primary_column_from_Text_string",
		mode: page.ArticlePrimaryColumnTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Text:  `<p>this primary-column text is from a string</p>`,
			}},
		}},
	}, {
		file: "primary_column_from_Text_file",
		mode: page.ArticlePrimaryColumnTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Text:  testFile,
			}},
		}},
	}, {
		file: "primary_column_from_Text_htmler",
		mode: page.ArticlePrimaryColumnTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Text:  testhtmler{},
			}},
		}},
	}, {
		file: "primary_column_from_Text_valuer",
		mode: page.ArticlePrimaryColumnTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Text:  testvaluer{httpdoc.V1{}},
			}},
		}},
	}, {
		file: "primary_column_from_Text_interface{}",
		mode: page.ArticlePrimaryColumnTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Text:  httpdoc.V1{},
			}},
		}},
	}, {
		file: "primary_column_from_Code_valuer",
		mode: page.ArticlePrimaryColumnTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Code:  testvaluer{httpdoc.V1{}},
			}},
		}},
	}, {
		file: "primary_column_from_Code_interface{}",
		mode: page.ArticlePrimaryColumnTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Code:  httpdoc.V1{},
			}},
		}},
	}, {
		///////////////////////////////////////////////////////////////
		// Article Code Column
		/////////////////////////////////////////////////////////////////
		file: "example_column_from_Code_string",
		mode: page.ArticleExampleColumnTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Code:  `<p>this example-column text is from a string</p>`,
			}},
		}},
	}, {
		file: "example_column_from_Code_file",
		mode: page.ArticleExampleColumnTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Code:  testFile,
			}},
		}},
	}, {
		file: "example_column_from_Code_htmler",
		mode: page.ArticleExampleColumnTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Code:  testhtmler{},
			}},
		}},
	}, {
		file: "example_column_from_Code_valuer",
		mode: page.ArticleExampleColumnTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Code:  testvaluer{httpdoc.V1{}},
			}},
		}},
	}, {
		file: "example_column_from_Code_interface{}",
		mode: page.ArticleExampleColumnTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Code:  httpdoc.V1{},
			}},
		}},
	}, {
		////////////////////////////////////////////////////////////////
		// auth_info_from_test_request
		////////////////////////////////////////////////////////////////
		file: "auth_info_from_test_request_auth_htmler",
		mode: page.ArticleAuthInfoTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Tests: []*httptest.Test{{
						Request: httptest.Request{
							Auth: testhtmler{},
						},
					}},
				}},
			}},
		}},
	}, {
		file: "auth_info_from_test_request_auth_valuer",
		mode: page.ArticleAuthInfoTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Tests: []*httptest.Test{{
						Request: httptest.Request{
							Auth: testvaluer{httpdoc.A1{}},
						},
					}},
				}},
			}},
		}},
	}, {
		////////////////////////////////////////////////////////////////
		// field_list_from_test_response
		////////////////////////////////////////////////////////////////
		file: "field_list_from_test_response_header",
		mode: page.ArticleFieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Tests: []*httptest.Test{{
						Response: httptest.Response{
							Header: testvaluer{httpdoc.H1RES{}},
						},
					}},
				}},
			}},
		}},
	}, {
		file: "field_list_from_test_response_body",
		mode: page.ArticleFieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Tests: []*httptest.Test{{
						Response: httptest.Response{
							Body: jsonbody{httpdoc.T1{}},
						},
					}},
				}},
			}},
		}},
	}, {
		file: "field_list_from_test_response",
		mode: page.ArticleFieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Tests: []*httptest.Test{{
						Response: httptest.Response{
							Header: testvaluer{httpdoc.H1RES{}},
							Body:   jsonbody{httpdoc.T1{}},
						},
					}},
				}},
			}},
		}},
	}, {
		////////////////////////////////////////////////////////////////
		// field_list_from_test_request
		////////////////////////////////////////////////////////////////
		file: "field_list_from_test_request_path",
		mode: page.ArticleFieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Tests: []*httptest.Test{{
						Request: httptest.Request{
							Params: testvaluer{httpdoc.P1{}},
						},
					}},
				}},
			}},
		}},
	}, {
		file: "field_list_from_test_request_query",
		mode: page.ArticleFieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Tests: []*httptest.Test{{
						Request: httptest.Request{
							Query: testvaluer{httpdoc.Q1{}},
						},
					}},
				}},
			}},
		}},
	}, {
		file: "field_list_from_test_request_header",
		mode: page.ArticleFieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Tests: []*httptest.Test{{
						Request: httptest.Request{
							Header: testvaluer{httpdoc.H1REQ{}},
						},
					}},
				}},
			}},
		}},
	}, {
		file: "field_list_from_test_request_body",
		mode: page.ArticleFieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Tests: []*httptest.Test{{
						Request: httptest.Request{
							Body: jsonbody{httpdoc.T1{}},
						},
					}},
				}},
			}},
		}},
	}, {
		file: "field_list_from_test_request",
		mode: page.ArticleFieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Tests: []*httptest.Test{{
						Request: httptest.Request{
							Params: testvaluer{httpdoc.P1{}},
							Query:  testvaluer{httpdoc.Q1{}},
							Header: testvaluer{httpdoc.H1REQ{}},
							Body:   jsonbody{httpdoc.T1{}},
						},
					}},
				}},
			}},
		}},
	}, {
		////////////////////////////////////////////////////////////////
		// field_list_from_test_request_and_response
		////////////////////////////////////////////////////////////////
		file: "field_list_from_test_request_and_response",
		mode: page.ArticleSectionListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Tests: []*httptest.Test{{
						Request: httptest.Request{
							Params: testvaluer{httpdoc.P1{}},
							Header: testvaluer{httpdoc.H1REQ{}},
						},
						Response: httptest.Response{
							Header: testvaluer{httpdoc.H1RES{}},
							Body:   jsonbody{httpdoc.T1{}},
						},
					}},
				}},
			}},
		}},
	}, {
		/////////////////////////////////////////////////////////////////
		// field_item_from_test_response
		/////////////////////////////////////////////////////////////////
		file: "field_item_from_test_response",
		mode: page.FieldItemTest,
		cfg: Config{
			SourceURL: srclink,
			FieldType: func(f reflect.StructField) (typeName string, ok bool) {
				return strings.ToUpper(f.Type.Kind().String()) + "_TYPE", true
			},
		},
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Tests: []*httptest.Test{{
						Response: httptest.Response{
							Body: jsonbody{httpdoc.FT1{}},
						},
					}},
				}},
			}},
		}},
	}, {
		file: "field_item_from_test_request",
		mode: page.FieldItemTest,
		cfg: Config{
			SourceURL: srclink,
			FieldType: func(f reflect.StructField) (typeName string, ok bool) {
				return strings.ToUpper(f.Type.Kind().String()) + "_TYPE", true
			},
			FieldSetting: func(s reflect.StructField, t reflect.Type) (label, text string, ok bool) {
				tag := tagutil.New(string(s.Tag))
				if tag.Contains("set", "required") {
					return "required", "IS_REQUIRED", true
				}
				return "", "", false
			},
			FieldValidation: func(s reflect.StructField, t reflect.Type) (text template.HTML) {
				tag := tagutil.New(string(s.Tag))
				if v := tag.First("validation"); len(v) > 0 {
					vv := strings.Split(v, ":")
					switch vv[0] {
					case "len":
						return template.HTML("<p>value must be of length between " +
							"<code>" + vv[1] + "</code> and <code>" + vv[2] +
							"</code> characters long</p>")
					}
				}
				return ""
			},
		},
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Tests: []*httptest.Test{{
						Request: httptest.Request{
							Body: jsonbody{httpdoc.FT1{}},
						},
					}},
				}},
			}},
		}},
	}, {
		/////////////////////////////////////////////////////////////////
		// Enum List
		/////////////////////////////////////////////////////////////////
		file: "enum_list_1",
		cfg:  Config{SourceURL: srclink},
		mode: page.EnumListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				Code:  httpdoc.T7{},
			}},
		}},
	}, {
		file: "enum_list_2",
		cfg:  Config{SourceURL: srclink},
		mode: page.EnumListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				Code:  httpdoc.T8{},
			}},
		}},
	}, {
		/////////////////////////////////////////////////////////////////
		// Endpoints Example
		/////////////////////////////////////////////////////////////////
		file: "example_endpoints",
		mode: page.ExampleEndpointsTest,
		toc: []*ArticleGroup{{
			Name: "Endpoint Group 1",
			Articles: []*Article{{
				Title: "Article 1",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"POST", "/api/foos"},
					Desc:     "Create a Foo",
				}, {
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Desc:     "List Foos",
				}, {
					Endpoint: httptest.Endpoint{"GET", "/api/foos/{id}"},
					Desc:     "Get a Foo",
				}, {
					Endpoint: httptest.Endpoint{"DELETE", "/api/foos"},
					Desc:     "Delete a Foo",
				}},
			}},
		}},
	}, {
		/////////////////////////////////////////////////////////////////
		// Response Example
		/////////////////////////////////////////////////////////////////
		file: "example_response",
		mode: page.ExampleResponseTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos/{id}"}, Desc: "Read Foo",
					Tests: []*httptest.Test{{
						Response: httptest.Response{
							StatusCode: 201,
							Body: jsonbody{httpdoc.T3{
								F1: "foo bar",
								F2: 0.007,
								F3: 12345,
								F4: true,
							}},
						},
					}},
				}},
			}},
		}},
	}, {
		file: "example_response_with_header",
		mode: page.ExampleResponseTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"GET", "/api/foos/{id}"}, Desc: "Read Foo",
					Tests: []*httptest.Test{{
						Response: httptest.Response{
							StatusCode: 200,
							Header: httptest.Header{
								"Content-Type": {"application/json"},
								"Set-Cookie":   {"s=11234567890", "t=0987654321"},
							},
							Body: jsonbody{httpdoc.T1{}},
						},
					}},
				}},
			}},
		}},
	}}

	for _, tt := range tests {
		if tt.skip {
			continue
		}

		t.Run(tt.file, func(t *testing.T) {
			defer func() {
				// move the file cursor back to the beginning
				// so it can be reused by multiple tests
				if _, err := testFile.Seek(0, 0); err != nil {
					t.Errorf("testFile.Seek(0, 0) fail: %v", err)
				}
			}()

			want, err := ioutil.ReadFile(testdatadir + "/httpdoc/" + tt.file + ".html")
			if err != nil {
				t.Error(err)
				return
			}

			tt.cfg.normalize()

			b := build{Config: tt.cfg, dir: tt.toc, mode: tt.mode}
			if err := b.loadCallerSource(0); err != nil {
				t.Error(err)
			} else if err := b.run(); err != nil {
				t.Error(err)
			} else {
				got := b.buf.Bytes()
				got, want = flatten(got), flatten(want)
				if e := compare.Compare(string(got), string(want)); e != nil {
					t.Error(e)
				}
			}
		})
	}
}

func flatten(data []byte) (out []byte) {
	out = make([]byte, len(data))
	skipNl, skipTab, j := true, true, 0
	for i := 0; i < len(data); i++ {
		if skipNl && data[i] == '\n' {
			skipTab = true
			continue
		}
		if skipTab && data[i] == '\t' {
			skipNl = false
			continue
		}

		skipNl = (data[i] == '\n')
		skipTab = (data[i] == '\n')

		out[j] = data[i]
		j += 1
	}

	return out[:j]
}

func findRootDir(wd string) (string, error) {
	dir, err := filepath.Abs(wd)
	if err != nil {
		return "", err
	}

	for len(dir) > 1 && dir[0] == '/' {
		if isRoot, err := isRootDir(dir); err != nil {
			return "", err
		} else if isRoot {
			return dir, nil
		}

		// parent dir will be examined next
		dir = filepath.Dir(dir)
	}

	return "", nil
}

// isRootDir reports if the directory at the given path is the root directory of a git project.
func isRootDir(path string) (bool, error) {
	d, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer d.Close()

	infoList, err := d.Readdir(-1)
	if err != nil {
		return false, err
	}

	for _, info := range infoList {
		name := info.Name()
		if name == ".git" && info.IsDir() {
			return true, nil
		}
	}

	return false, nil
}

type testhtmler struct{}

func (testhtmler) HTML() (HTML, error) {
	return `<p>this text is from an httpdoc.HTMLer</p>`, nil
}
func (v testhtmler) SetAuth(r *http.Request) {}

type testvaluer struct{ v interface{} }

func (v testvaluer) Value() (Value, error)           { return v.v, nil }
func (v testvaluer) GetHeader() http.Header          { return nil }
func (v testvaluer) SetParams(pattern string) string { return "" }
func (v testvaluer) GetQuery() string                { return "" }
func (v testvaluer) SetAuth(r *http.Request)         {}

type jsonbody struct{ v interface{} }

func (b jsonbody) Value() (Value, error)     { return b.v, nil }
func (b jsonbody) Type() string              { return "application/json" }
func (b jsonbody) Compare(r io.Reader) error { return nil }

func (b jsonbody) Reader() (io.Reader, error) {
	bs, err := json.Marshal(b.v)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(bs), nil
}
