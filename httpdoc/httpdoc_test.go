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
	"github.com/frk/httptest/internal/testdata/httpdoc"
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
		toc             []*TopicGroup
	}{{
		///////////////////////////////////////////////////////////////
		// Sidebar
		/////////////////////////////////////////////////////////////////
		file: "sidebar_from_topics",
		mode: page.SidebarTest,
		toc: []*TopicGroup{{
			Name: "Topic Group 1",
			Topics: []*Topic{{
				Name: "Topic 1",
				SubTopics: []*Topic{{
					Name: "Sub Topic 1",
				}, {
					Name: "Sub Topic 2",
					SubTopics: []*Topic{{
						Name: "Sub Sub Topic 1",
					}},
				}},
			}, {
				Name: "Topic 2",
			}},
		}, {
			Name: "Topic Group 2",
			Topics: []*Topic{{
				Name: "Topic 3",
			}},
		}},
	}, {
		file: "sidebar_from_endpoints",
		mode: page.SidebarTest,
		toc: []*TopicGroup{{
			Name: "Endpoint Group 1",
			Topics: []*Topic{{
				Name: "Topic 1",
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
		toc: []*TopicGroup{{
			Name: "Endpoint Group 1",
			Topics: []*Topic{{
				Name: "Topic 1",
				TestGroups: []*httptest.TestGroup{{
					Endpoint: httptest.Endpoint{"POST", "/api/foos"},
					Desc:     "Create a Foo",
				}, {
					Endpoint: httptest.Endpoint{"GET", "/api/foos"},
					Desc:     "List Foos",
				}},
				SubTopics: []*Topic{{
					Name: "Sub Topic 1",
				}, {
					Name: "Sub Topic 2",
					SubTopics: []*Topic{{
						Name: "Sub Sub Endpoints 1",
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
			Name: "Topic Group 2",
			Topics: []*Topic{{
				Name: "Topic 3",
			}},
		}},
	}, {
		///////////////////////////////////////////////////////////////
		// Content
		/////////////////////////////////////////////////////////////////
		file: "content_from_topics",
		mode: page.ContentTest,
		toc: []*TopicGroup{{
			Name: "Topic Group 1",
			Topics: []*Topic{{
				Name: "Topic 1",
				SubTopics: []*Topic{{
					Name: "Sub Topic 1",
				}, {
					Name: "Sub Topic 2",
					SubTopics: []*Topic{{
						Name: "Sub Sub Topic 1",
					}},
				}},
			}, {
				Name: "Topic 2",
			}},
		}, {
			Name: "Topic Group 2",
			Topics: []*Topic{{
				Name: "Topic 3",
			}},
		}},
	}, {
		///////////////////////////////////////////////////////////////
		// Article Text
		/////////////////////////////////////////////////////////////////
		file: "article_from_topic_raw_string",
		mode: page.ArticleTest,
		toc: []*TopicGroup{{
			Name: "Topic Group 1",
			Topics: []*Topic{{
				Name: "Article",
				Doc: `<div>
				<h4>Test</h4>
				<p>this is a raw string</p>
				</div>`, //`
			}},
		}},
	}, {
		file: "article_from_topic_file",
		mode: page.ArticleTest,
		toc: []*TopicGroup{{
			Name: "Topic Group 1",
			Topics: []*Topic{{
				Name: "Article",
				Doc:  testFile,
			}},
		}},
	}, {
		///////////////////////////////////////////////////////////////
		// Article "Returns"
		/////////////////////////////////////////////////////////////////
		file: "article_conclusion_from_topic_raw_string",
		mode: page.ArticleTest,
		toc: []*TopicGroup{{
			Name: "Topic Group 1",
			Topics: []*Topic{{
				Name: "Article",
				Doc:  `<p>this is a test article</p>`,
				Returns: `<div>
				<h4>Test</h4>
				<p>this is a raw string</p>
				</div>`, //`
			}},
		}},
	}, {
		file: "article_conclusion_from_topic_file",
		mode: page.ArticleTest,
		toc: []*TopicGroup{{
			Name: "Topic Group 1",
			Topics: []*Topic{{
				Name:    "Article",
				Doc:     `<p>this is a test article</p>`,
				Returns: testFile,
			}},
		}},
	}, {
		/////////////////////////////////////////////////////////////////
		// Article Fields (Attributes)
		/////////////////////////////////////////////////////////////////
		file: "article_field_list_attributes",
		mode: page.FieldListTest,
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Attributes: httpdoc.T1{},
			}},
		}},
	}, {
		file: "article_field_list_attributes_with_comments",
		mode: page.FieldListTest,
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Attributes: httpdoc.T2{},
			}},
		}},
	}, {
		file: "article_field_list_attributes_with_tag_names",
		mode: page.FieldListTest,
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Attributes: httpdoc.T3{},
			}},
		}},
	}, {
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
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Attributes: httpdoc.T1{},
			}},
		}},
	}, {
		file: "article_field_list_attributes_with_nested_fields_1",
		mode: page.FieldListTest,
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Attributes: httpdoc.T4{},
			}},
		}},
	}, {
		file: "article_field_list_attributes_with_nested_fields_2",
		mode: page.FieldListTest,
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Attributes: httpdoc.T5{},
			}},
		}},
	}, {
		file:    "article_field_list_attributes_with_nested_fields_3",
		rootdir: rootdir,
		repourl: repourl,
		mode:    page.FieldListTest,
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Attributes: httpdoc.T6{},
			}},
		}},
	}, {
		////////////////////////////////////////////////////////////////
		// Article Fields (Parameters)
		////////////////////////////////////////////////////////////////
		file: "article_field_list_parameters_1",
		mode: page.FieldListTest,
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Parameters: httpdoc.T1{},
			}},
		}},
	}, {
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
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Parameters: httpdoc.T3{},
			}},
		}},
	}, {
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
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Parameters: httpdoc.T3{},
			}},
		}},
	}, {
		/////////////////////////////////////////////////////////////////
		// Enum List
		/////////////////////////////////////////////////////////////////
		file:    "field_enum_list",
		rootdir: rootdir,
		repourl: repourl,
		mode:    page.FieldItemTest,
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Attributes: httpdoc.T7{},
			}},
		}},
	}, {
		/////////////////////////////////////////////////////////////////
		// Example Endpoint Overview
		/////////////////////////////////////////////////////////////////
		file: "endpoint_overview",
		mode: page.EndpointOverviewTest,
		toc: []*TopicGroup{{
			Name: "Endpoint Group 1",
			Topics: []*Topic{{
				Name: "Topic 1",
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
