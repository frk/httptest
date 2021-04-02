package httpdoc

import (
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/frk/compare"
	"github.com/frk/httptest"
	"github.com/frk/httptest/internal/page"
	"github.com/frk/httptest/internal/testdata/httpdoc"
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
		file    string
		rootdir string
		repourl string
		typName func(reflect.StructField) (typeName string, ok bool)
		mode    page.TestMode
		toc     []*TopicGroup
	}{{
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
		file: "article_from_topic_template_html",
		mode: page.ArticleTest,
		toc: []*TopicGroup{{
			Name: "Topic Group 1",
			Topics: []*Topic{{
				Name: "Article",
				Text: template.HTML(`<div>
				<h4>Test</h4>
				<p>this is template.HTML</p>
				</div>`), //`
			}},
		}},
	}, {
		file: "article_from_topic_raw_string",
		mode: page.ArticleTest,
		toc: []*TopicGroup{{
			Name: "Topic Group 1",
			Topics: []*Topic{{
				Name: "Article",
				Text: `<div>
				<h4>Test</h4>
				<p>this is a raw string</p>
				</div>`, //`
			}},
		}},
	}, {
		file: "article_from_topic_html_iface",
		mode: page.ArticleTest,
		toc: []*TopicGroup{{
			Name: "Topic Group 1",
			Topics: []*Topic{{
				Name: "Article",
				Text: htmlimpl{str: `<div>
				<h4>Test</h4>
				<p>this is from httpdoc.HTML.HTML()</p>
				</div>`}, //`
			}},
		}},
	}, {
		file: "article_from_topic_file",
		mode: page.ArticleTest,
		toc: []*TopicGroup{{
			Name: "Topic Group 1",
			Topics: []*Topic{{
				Name: "Article",
				Text: testFile,
			}},
		}},
	}, {
		file: "article_field_list_attributes",
		mode: page.ArticleFieldListTest,
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Attributes: httpdoc.T1{},
			}},
		}},
	}, {
		file: "article_field_list_attributes_with_comments",
		mode: page.ArticleFieldListTest,
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Attributes: httpdoc.T2{},
			}},
		}},
	}, {
		file: "article_field_list_attributes_with_tag_names",
		mode: page.ArticleFieldListTest,
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Attributes: httpdoc.T3{},
			}},
		}},
	}, {
		file: "article_field_list_attributes_with_custom_type_names",
		mode: page.ArticleFieldListTest,
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
		mode: page.ArticleFieldListTest,
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Attributes: httpdoc.T4{},
			}},
		}},
	}, {
		file: "article_field_list_attributes_with_nested_fields_2",
		mode: page.ArticleFieldListTest,
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
		mode:    page.ArticleFieldListTest,
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Attributes: httpdoc.T6{},
			}},
		}},
	}, {
		file:    "field_enum_list",
		rootdir: rootdir,
		repourl: repourl,
		mode:    page.ArticleFieldItemTest,
		toc: []*TopicGroup{{
			Topics: []*Topic{{
				Name:       "Test Topic",
				Attributes: httpdoc.T7{},
			}},
		}},
	}, {
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
		t.Run(tt.file, func(t *testing.T) {
			want, err := ioutil.ReadFile(testdatadir + "/httpdoc/" + tt.file + ".html")
			if err != nil {
				t.Error(err)
				return
			}

			c := Config{
				ProjectRoot:   tt.rootdir,
				RepositoryURL: tt.repourl,
				FieldTypeName: tt.typName,
				mode:          tt.mode,
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

type htmlimpl struct {
	str string
	err error
}

func (h htmlimpl) HTML() (template.HTML, error) {
	if h.err != nil {
		return "", h.err
	}
	return template.HTML(h.str), nil
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
