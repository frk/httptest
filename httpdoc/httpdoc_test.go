package httpdoc

import (
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/frk/compare"
	"github.com/frk/httptest"
	"github.com/frk/httptest/internal/page"
	//"github.com/frk/httptest/internal/testdata/httpdoc"
	"github.com/frk/tagutil"
)

func Test(t *testing.T) {
	_, f, _, _ := runtime.Caller(0)
	workdir := filepath.Dir(f)
	rootdir, err := findRootDir(workdir)
	if err != nil {
		t.Error(err)
		return
	}

	repourl := "https://github.com/frk/httptest/tree/master/"

	testdatadir := filepath.Dir(workdir) + "/internal/testdata"
	testFile, err := os.Open(testdatadir + "/test.html")
	if err != nil {
		t.Error(err)
		return
	}
	defer testFile.Close()

	tests := []struct {
		skip            bool
		file            string
		rootdir         string
		repourl         string
		typName         func(reflect.StructField) (typeName string, ok bool)
		fieldSetting    func(reflect.StructField, reflect.Type) (label, text string, ok bool)
		fieldValidation func(reflect.StructField, reflect.Type) (text template.HTML)
		mode            page.TestMode
		toc             []*ArticleGroup
	}{{
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
		///////////////////////////////////////////////////////////////
		// Article Text
		/////////////////////////////////////////////////////////////////
		file: "article_text_from_raw_string",
		mode: page.ArticleTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Text: `<div>
				<h4>Test</h4>
				<p>this is a raw string</p>
				</div>`, //`
			}},
		}},
	}, {
		skip: true,
		file: "article_from_topic_file",
		mode: page.ArticleTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Text:  testFile,
			}},
		}},
	}, {
		///////////////////////////////////////////////////////////////
		// Article "Returns"
		/////////////////////////////////////////////////////////////////
		skip: true,
		file: "article_conclusion_from_topic_raw_string",
		mode: page.ArticleTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Text:  `<p>this is a test article</p>`,
				//Returns: `<div>
				//<h4>Test</h4>
				//<p>this is a raw string</p>
				//</div>`, //`
			}},
		}},
	}, {
		skip: true,
		file: "article_conclusion_from_topic_file",
		mode: page.ArticleTest,
		toc: []*ArticleGroup{{
			Name: "Article Group 1",
			Articles: []*Article{{
				Title: "Article",
				Text:  `<p>this is a test article</p>`,
				//Returns: testFile,
			}},
		}},
	}, {
		/////////////////////////////////////////////////////////////////
		// Article Fields (Attributes)
		/////////////////////////////////////////////////////////////////
		skip: true,
		file: "article_field_list_attributes",
		mode: page.FieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				//Attributes: httpdoc.T1{},
			}},
		}},
	}, {
		skip: true,
		file: "article_field_list_attributes_with_comments",
		mode: page.FieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				//Attributes: httpdoc.T2{},
			}},
		}},
	}, {
		skip: true,
		file: "article_field_list_attributes_with_tag_names",
		mode: page.FieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				//Attributes: httpdoc.T3{},
			}},
		}},
	}, {
		skip: true,
		file: "article_field_list_attributes_with_custom_type_names",
		mode: page.FieldListTest,
		typName: func(f reflect.StructField) (typeName string, ok bool) {
			name := []byte(f.Type.String())
			i, j := 0, len(name)-1
			for i < j {
				name[i], name[j] = name[j], name[i]
				i, j = i+1, j-1
			}
			return string(name), true
		},
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				//Attributes: httpdoc.T1{},
			}},
		}},
	}, {
		skip: true,
		file: "article_field_list_attributes_with_nested_fields_1",
		mode: page.FieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				//Attributes: httpdoc.T4{},
			}},
		}},
	}, {
		skip: true,
		file: "article_field_list_attributes_with_nested_fields_2",
		mode: page.FieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				//Attributes: httpdoc.T5{},
			}},
		}},
	}, {
		skip:    true,
		file:    "article_field_list_attributes_with_nested_fields_3",
		rootdir: rootdir,
		repourl: repourl,
		mode:    page.FieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				//Attributes: httpdoc.T6{},
			}},
		}},
	}, {
		////////////////////////////////////////////////////////////////
		// Article Fields (Parameters)
		////////////////////////////////////////////////////////////////
		skip: true,
		file: "article_field_list_parameters_1",
		mode: page.FieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				//Parameters: httpdoc.T1{},
			}},
		}},
	}, {
		skip: true,
		file: "article_field_list_parameters_2",
		fieldSetting: func(s reflect.StructField, t reflect.Type) (label, text string, ok bool) {
			tag := tagutil.New(string(s.Tag))
			if tag.Contains("set", "required") {
				return "required", "This field is required", true
			}
			if tag.Contains("set", "optional") {
				return "optional", "This field is optional", true
			}
			if tag.Contains("set", "conditional") {
				return "conditional", "This field is conditional", true
			}
			return "", "", false
		},
		mode: page.FieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				//Parameters: httpdoc.T3{},
			}},
		}},
	}, {
		skip: true,
		file: "article_field_list_parameters_3",
		fieldValidation: func(s reflect.StructField, t reflect.Type) (text template.HTML) {
			tag := tagutil.New(string(s.Tag))
			if v := tag.First("validation"); len(v) > 0 {
				vv := strings.Split(v, ":")
				switch vv[0] {
				case "len":
					return template.HTML("<p>value must be of length between " +
						"<code>" + vv[1] + "</code> and <code>" + vv[2] +
						"</code> characters long</p>")
				case "min":
					return template.HTML("<p>value must be at least " +
						"<code>" + vv[1] + "</code></p>")
				case "max":
					return template.HTML("<p>value must be at most " +
						"<code>" + vv[1] + "</code></p>")
				}
			}
			return ""
		},
		mode: page.FieldListTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				//Parameters: httpdoc.T3{},
			}},
		}},
	}, {
		/////////////////////////////////////////////////////////////////
		// Enum List
		/////////////////////////////////////////////////////////////////
		skip:    true,
		file:    "field_enum_list",
		rootdir: rootdir,
		repourl: repourl,
		mode:    page.FieldItemTest,
		toc: []*ArticleGroup{{
			Articles: []*Article{{
				Title: "Test Article",
				//Attributes: httpdoc.T7{},
			}},
		}},
	}, {
		/////////////////////////////////////////////////////////////////
		// Example Endpoint Overview
		/////////////////////////////////////////////////////////////////
		skip: true,
		file: "endpoint_overview",
		mode: page.EndpointOverviewTest,
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

			c := Config{
				ProjectRoot:     tt.rootdir,
				RepositoryURL:   tt.repourl,
				FieldType:       tt.typName,
				FieldSetting:    tt.fieldSetting,
				FieldValidation: tt.fieldValidation,
				mode:            tt.mode,
			}
			if err := c.Build(tt.toc); err != nil {
				t.Error(err)
			} else {
				got := c.buf.Bytes()
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
