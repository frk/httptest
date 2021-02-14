package integration

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/frk/compare"
	"github.com/frk/httptest/httpdoc"
	"github.com/frk/httptest/httpdoc/internal/testdata/in"
)

func Test(t *testing.T) {
	_, f, _, _ := runtime.Caller(0)
	outdir := filepath.Dir(filepath.Dir(f)) + "/testdata/out"

	tests := []struct {
		file string
		toc  []*httpdoc.TopicGroup
		// ...
	}{
		{"topic_one", in.TopicOne},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			want, err := ioutil.ReadFile(outdir + "/" + tt.file + ".html")
			if err != nil {
				t.Error(err)
				return
			}

			var c httpdoc.Config
			c.Build(tt.toc)
			got := c.Bytes()

			if e := compare.Compare(string(got), string(want)); e != nil {
				t.Error(e)
			}
		})
	}
}
