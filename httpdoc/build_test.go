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

var test_file_dir string
var test_data_dir string

func init() {
	_, f, _, _ := runtime.Caller(0)
	test_file_dir = filepath.Dir(f)
	test_data_dir = filepath.Dir(test_file_dir) + "/internal/testdata"
}

func Test_build(t *testing.T) {
	defer closefiles()

	srclocal, err := findRootDir(test_file_dir)
	if err != nil {
		t.Error(err)
		return
	}

	srclink := testSourceURLFunc(srclocal)

	tests := []struct {
		skip bool
		file string
		cfg  Config
		wt   writeTest
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
			Name:         "Group Name",
			Articles:     []*Article{{Title: "Article Title"}},
			LoadExpanded: true,
		}},
	}, {
		file: "page_with_endpoint",
		toc: []*ArticleGroup{{
			Name: "Group Name",
			Articles: []*Article{{
				Title: "Article Title",
				TestGroups: []*httptest.TestGroup{{
					E: "POST /api/foos",
				}},
			}},
			LoadExpanded: true,
		}},
	}, {
		///////////////////////////////////////////////////////////////
		// Sidebar
		/////////////////////////////////////////////////////////////////
		file: "sidebar_from_articles",
		wt:   wt_sidebar,
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
		wt:   wt_sidebar,
		toc: []*ArticleGroup{{
			Name: "Endpoint Group 1",
			Articles: []*Article{{
				Title: "Article 1",
				TestGroups: []*httptest.TestGroup{{
					E:    "POST /api/foos",
					Desc: "Create a Foo",
				}, {
					E:    "GET /api/foos",
					Desc: "List Foos",
				}, {
					E:    "GET /api/foos/{id}",
					Desc: "Get a Foo",
				}, {
					E:    "DELETE /api/foos",
					Desc: "Delete a Foo",
				}},
			}},
		}},
	}, {
		file: "sidebar_from_mix",
		wt:   wt_sidebar,
		toc: []*ArticleGroup{{
			Name: "Endpoint Group 1",
			Articles: []*Article{{
				Title: "Article 1",
				TestGroups: []*httptest.TestGroup{{
					E:    "POST /api/foos",
					Desc: "Create a Foo",
				}, {
					E:    "GET /api/foos",
					Desc: "List Foos",
				}},
				SubArticles: []*Article{{
					Title: "Sub Article 1",
				}, {
					Title: "Sub Article 2",
					SubArticles: []*Article{{
						Title: "Sub Sub Endpoints 1",
						TestGroups: []*httptest.TestGroup{{
							E:    "POST /api/foos/{id}/bars",
							Desc: "Create a FooBar",
						}, {
							E:    "GET /api/foos/{id}/bars",
							Desc: "List FooBars",
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
		// Sidebar Header
		/////////////////////////////////////////////////////////////////
		file: "sidebar_header_banner_from_page_title",
		wt:   wt_sidebar_header,
		cfg: Config{
			PageTitle: "The Page's Title",
		},
		toc: []*ArticleGroup{},
	}, {
		file: "sidebar_header_banner_from_html",
		wt:   wt_sidebar_header,
		cfg: Config{
			SidebarBannerHTML: "<h1>The Page's Banner</h1>",
		},
		toc: []*ArticleGroup{},
	}, {
		///////////////////////////////////////////////////////////////
		// Content
		/////////////////////////////////////////////////////////////////
		file: "content_from_articles",
		wt:   wt_content,
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
			LoadExpanded: true,
		}, {
			Name: "Article Group 2",
			Articles: []*Article{{
				Title: "Article 3",
			}},
			LoadExpanded: true,
		}},
	}, {
		file: "content_from_endpoints",
		wt:   wt_content,
		toc: []*ArticleGroup{{
			Name: "Endpoint Group 1",
			Articles: []*Article{{
				Title: "Article 1",
				TestGroups: []*httptest.TestGroup{{
					E:    "POST /api/foos",
					Desc: "Create a Foo",
				}, {
					E:    "GET /api/foos",
					Desc: "List Foos",
				}, {
					E:    "GET /api/foos/{id}",
					Desc: "Get a Foo",
				}, {
					E:    "DELETE /api/foos",
					Desc: "Delete a Foo",
				}},
			}},
			LoadExpanded: true,
		}},
	}, {
		file: "content_from_mix",
		wt:   wt_content,
		toc: []*ArticleGroup{{
			Name: "Endpoint Group 1",
			Articles: []*Article{{
				Title: "Article 1",
				TestGroups: []*httptest.TestGroup{{
					E:    "POST /api/foos",
					Desc: "Create a Foo",
				}, {
					E:    "GET /api/foos",
					Desc: "List Foos",
				}},
				SubArticles: []*Article{{
					Title: "Sub Article 1",
				}, {
					Title: "Sub Article 2",
					SubArticles: []*Article{{
						Title: "Sub Sub Endpoints 1",
						TestGroups: []*httptest.TestGroup{{
							E:    "GET /api/foos/{id}/bars",
							Desc: "List FooBars",
						}},
					}},
				}},
			}},
			LoadExpanded: true,
		}, {
			Name: "Article Group 2",
			Articles: []*Article{{
				Title: "Article 3",
			}},
			LoadExpanded: true,
		}},
	}, {
		///////////////////////////////////////////////////////////////
		// Article Text Column
		/////////////////////////////////////////////////////////////////
		file: "primary_column_from_Text_string",
		wt:   wt_article_primary_column,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Text:  `<p>this primary-column text is from a string</p>`,
			}},
		}},
	}, {
		file: "primary_column_from_Text_file",
		wt:   wt_article_primary_column,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Text:  openfile("test.html"),
			}},
		}},
	}, {
		file: "primary_column_from_Text_htmler",
		wt:   wt_article_primary_column,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Text:  testhtmler{},
			}},
		}},
	}, {
		file: "primary_column_from_Text_valuer",
		wt:   wt_article_primary_column,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Text:  testvaluer{httpdoc.V1{}},
			}},
		}},
	}, {
		file: "primary_column_from_Text_interface{}",
		wt:   wt_article_primary_column,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Text:  httpdoc.V1{},
			}},
		}},
	}, {
		file: "primary_column_from_Code_valuer",
		wt:   wt_article_primary_column,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Code:  testvaluer{httpdoc.V1{}},
			}},
		}},
	}, {
		file: "primary_column_from_Code_interface{}",
		wt:   wt_article_primary_column,
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
		wt:   wt_article_example_column,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Code:  `<p>this example-column text is from a string</p>`,
			}},
		}},
	}, {
		file: "example_column_from_Code_file",
		wt:   wt_article_example_column,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Code:  openfile("test.html"),
			}},
		}},
	}, {
		file: "example_column_from_Code_htmler",
		wt:   wt_article_example_column,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Code:  testhtmler{},
			}},
		}},
	}, {
		file: "example_column_from_Code_valuer",
		wt:   wt_article_example_column,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Code:  testvaluer{httpdoc.V1{}},
			}},
		}},
	}, {
		file: "example_column_from_Code_interface{}",
		wt:   wt_article_example_column,
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
		wt:   wt_article_auth_info,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "GET /api/foos",
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
		wt:   wt_article_auth_info,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "GET /api/foos",
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
		wt:   wt_article_field_list,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "GET /api/foos",
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
		wt:   wt_article_field_list,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "GET /api/foos",
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
		wt:   wt_article_field_list,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "GET /api/foos",
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
		wt:   wt_article_field_list,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "GET /api/foos",
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
		wt:   wt_article_field_list,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "GET /api/foos",
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
		wt:   wt_article_field_list,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "GET /api/foos",
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
		wt:   wt_article_field_list,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "GET /api/foos",
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
		wt:   wt_article_field_list,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "GET /api/foos",
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
		wt:   wt_article_section_list,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "GET /api/foos",
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
		wt:   wt_field_item,
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
					E: "GET /api/foos",
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
		wt:   wt_field_item,
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
					E: "GET /api/foos",
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
		wt:   wt_enum_list,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				Code:  httpdoc.T7{},
			}},
		}},
	}, {
		file: "enum_list_2",
		cfg:  Config{SourceURL: srclink},
		wt:   wt_enum_list,
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
		wt:   wt_example_endpoints,
		toc: []*ArticleGroup{{
			Name: "Endpoint Group 1",
			Articles: []*Article{{
				Title: "Article 1",
				TestGroups: []*httptest.TestGroup{{
					E:    "POST /foos",
					Desc: "Create a Foo",
				}, {
					E:    "GET /foos",
					Desc: "List Foos",
				}, {
					E:    "GET /foos/{id}",
					Desc: "Get a Foo",
				}, {
					E:    "DELETE /foos",
					Desc: "Delete a Foo",
				}},
			}},
		}},
	}, {
		/////////////////////////////////////////////////////////////////
		// example_response
		/////////////////////////////////////////////////////////////////
		file: "example_response",
		wt:   wt_example_response,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "GET /api/foos/{id}", Desc: "Read Foo",
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
		wt:   wt_example_response,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "GET /api/foos/{id}", Desc: "Read Foo",
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
	}, {
		/////////////////////////////////////////////////////////////////
		// example_request
		/////////////////////////////////////////////////////////////////
		file: "example_request_topbar",
		wt:   wt_example_request_topbar,
		cfg:  Config{SnippetTypes: []SnippetType{SNIPP_HTTP, SNIPP_CURL}},
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E:     "POST /api/foos",
					Tests: []*httptest.Test{{}},
				}},
			}},
		}},
	}, {
		file: "example_request_body_http",
		wt:   wt_example_request_body,
		cfg:  Config{SnippetTypes: []SnippetType{SNIPP_HTTP}},
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "POST /api/foos",
					Tests: []*httptest.Test{{
						Request: httptest.Request{
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
		file: "example_request_body_curl",
		wt:   wt_example_request_body,
		cfg:  Config{SnippetTypes: []SnippetType{SNIPP_CURL}},
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				TestGroups: []*httptest.TestGroup{{
					E: "POST /api/foos",
					Tests: []*httptest.Test{{
						Request: httptest.Request{
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
	}}

	for _, tt := range tests {
		if tt.skip {
			continue
		}

		t.Run(tt.file, func(t *testing.T) {

			want, err := ioutil.ReadFile(test_data_dir + "/httpdoc/" + tt.file + ".html")
			if err != nil {
				t.Error(err)
				return
			}

			// build
			tt.cfg.srcdir = test_file_dir
			tt.cfg.normalize()
			b := build{Config: tt.cfg, dir: tt.toc}
			if err := b.run(); err != nil {
				t.Error(err)
				return
			}

			// write
			var buf bytes.Buffer
			if err := write_wt(&buf, b.page, tt.wt); err != nil {
				t.Error(err)
			} else {
				got := buf.Bytes()
				got, want = flatten(got), flatten(want)
				if e := compare.Compare(string(got), string(want)); e != nil {
					t.Error(e)
				}
			}
		})
	}
}

func flatten(data []byte) (out []byte) {
	preL := bytes.Index(data, []byte("<pre"))
	preR := bytes.Index(data, []byte("</pre>"))

	out = make([]byte, len(data))
	skipNl, skipTab, j := true, true, 0
	for i := 0; i < len(data); i++ {
		// keep <pre> formatting
		if preL > -1 {
			if i > preL && i < preR {
				out[j] = data[i]
				j += 1
				continue
			}
			if i >= preR {
				preL = bytes.Index(data[i:], []byte("<pre"))
				preR = bytes.Index(data[i:], []byte("</pre>"))
			}
		}

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

var openfiles []*os.File

func openfile(filename string) *os.File {
	filename = filepath.Join(test_data_dir, filename)
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	openfiles = append(openfiles, f)
	return f
}

func closefiles() {
	for _, f := range openfiles {
		f.Close()
	}
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

func testSourceURLFunc(local string) (f func(filename string, line int) (url string)) {
	if l := len(local); l > 0 && local[l-1] == '/' {
		local = local[:l-1]
	}

	remote := "https://github.com/frk/httptest/tree/master"
	return func(filename string, line int) (url string) {
		file := strings.TrimPrefix(filename, local)
		href := remote + file + "#L7357"
		return href
	}
}

type writeTest string

const (
	wt_sidebar        writeTest = "sidebar"
	wt_sidebar_header writeTest = "sidebar_header"

	wt_content writeTest = "content"

	wt_article_primary_column writeTest = "article_primary_column"
	wt_article_example_column writeTest = "article_example_column"

	wt_article_section_list writeTest = "article_section_list"
	wt_article_auth_info    writeTest = "article_auth_info"
	wt_article_field_list   writeTest = "article_field_list"

	wt_example_endpoints      writeTest = "example_endpoints"
	wt_example_response       writeTest = "example_response"
	wt_example_request_topbar writeTest = "example_request_topbar"
	wt_example_request_body   writeTest = "example_request_body"

	wt_field_item writeTest = "field_item"
	wt_enum_list  writeTest = "enum_list"
)

// It is expected that the part of the Page identified by the writeTest
// flag has been properly initialized by the test, if not the program may crash.
func write_wt(w io.Writer, p page.Page, wt writeTest) error {
	name := string(wt)
	data := interface{}(p)

	switch wt {
	case wt_sidebar:
		data = p.Sidebar
	case wt_sidebar_header:
		data = p.Sidebar.Header
	case wt_content:
		data = p.Content
	case wt_article_primary_column, wt_article_example_column:
		data = p.Content.Articles[0]

	// article section tests
	case wt_article_section_list:
		data = p.Content.Articles[0].SubArticles[0].Sections
	case wt_article_auth_info:
		data = p.Content.Articles[0].SubArticles[0].Sections[0].(*page.ArticleAuthInfo)
	case wt_article_field_list:
		data = p.Content.Articles[0].SubArticles[0].Sections[0].(*page.ArticleFieldList)

	// example section tests
	case wt_example_endpoints:
		data = p.Content.Articles[0].Example.Sections[0].(*page.ExampleEndpoints)
	case wt_example_response:
		data = p.Content.Articles[0].SubArticles[0].Example.Sections[1].(*page.ExampleResponse)
	case wt_example_request_topbar, wt_example_request_body:
		data = p.Content.Articles[0].SubArticles[0].Example.Sections[0].(*page.ExampleRequest)

	// single item tests
	case wt_field_item:
		data = p.Content.Articles[0].SubArticles[0].Sections[0].(*page.ArticleFieldList).Lists[0].Items[0]
	case wt_enum_list:
		data = p.Content.Articles[0].Sections[0].(*page.ArticleFieldList).Lists[0].Items[0].EnumList
	default:
		return page.T.Execute(w, p)
	}

	return page.T.ExecuteTemplate(w, name, data)
}
