package httpdoc

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/frk/httptest/internal/page"
	"github.com/frk/httptest/internal/program"
)

//
//
// mydoc
// ├── mydoc (executable)
// └── files
//     ├── html
//     │   ├── article-1.html
//     │   └── article-2.html
//     ├── css
//     │   └── main.css
//     └── js
//         └── main.js
//
//

// holds the state of the write
type write struct {
	Config
	page page.Page
	prog program.Program

	dperm    os.FileMode
	fperm    os.FileMode
	outdir   string
	tempdir  string
	filesdir string
	htmldir  string
	cssdir   string
	jsdir    string

	assetsdir string
}

func (w *write) run() error {
	defer func() {
		os.RemoveAll(w.tempdir)
	}()

	if err := w.init(); err != nil {
		return err
	}
	if err := w.writeArticles(); err != nil {
		return err
	}
	if err := w.writeProgram(); err != nil {
		return err
	}
	if err := w.copyAssets(); err != nil {
		return err
	}

	// remove old, if exits
	if err := os.RemoveAll(w.outdir); err != nil {
		return err
	}
	return os.Rename(w.tempdir, w.outdir)
}

func (w *write) init() error {
	w.outdir = filepath.Join(w.Config.OutputDir, w.Config.OutputName)

	// use the destination dir's file mode for the rest of the files
	fi, err := os.Stat(w.Config.OutputDir)
	if err != nil {
		return err
	}
	w.dperm = fi.Mode().Perm()
	w.fperm = w.dperm &^ 0111

	// initialize directory structure
	tempdir, err := ioutil.TempDir(w.OutputDir, "httpdoc_*")
	if err != nil {
		return err
	}

	w.tempdir = tempdir
	w.filesdir = filepath.Join(w.tempdir, "files")
	w.htmldir = filepath.Join(w.filesdir, "html")
	w.cssdir = filepath.Join(w.filesdir, "css")
	w.jsdir = filepath.Join(w.filesdir, "js")
	dirs := []string{w.filesdir, w.htmldir, w.cssdir, w.jsdir}
	for _, d := range dirs {
		if err := os.Mkdir(d, w.dperm); err != nil {
			return err
		}
	}

	// The w.pkgdir value is expected to hold this package's directory which is
	// rooted in the httptest module, the assets live in httptest/internal/assets/,
	// so get the parent httptest/ dir and go from there.
	//
	// NOTE(mkopriva): since this code relies on the directory structure
	// of the httptest module, it is important to keep in mind that if that
	// structure changes then this code will need to be changed as well.
	w.assetsdir = filepath.Join(filepath.Dir(w.pkgdir), "internal/assets")
	if f, err := os.Stat(w.assetsdir); err != nil || !f.IsDir() {
		return fmt.Errorf("can't find assets at: %q err: %w", w.assetsdir, err)
	}

	return nil
}

func (w *write) writeArticles() error {
	// Split the page into as many articles as there are elements in the
	// root p.Content.Articles slice, then write each into their own file.
	for _, aElem := range w.page.Content.Articles {
		p := w.page // shallow copy
		p.Content.Articles = []*page.ArticleElement{aElem}

		filename := filepath.Join(w.htmldir, aElem.Id+".html")
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, w.fperm)
		if err != nil {
			fmt.Println("os.OpenFile")
			return err
		}
		defer f.Close()

		if err := page.T.Execute(f, p); err != nil {
			fmt.Println("page.T.Execute")
			return err
		}
		if err := f.Sync(); err != nil {
			fmt.Println("f.Sync")
			return err
		}
	}
	return nil
}

func (w *write) writeProgram() error {
	// create the go source file for the program and write the contents to it
	srcfile := filepath.Join(w.tempdir, w.Config.OutputName+".go")
	f, err := os.OpenFile(srcfile, os.O_CREATE|os.O_WRONLY, w.fperm)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := program.T.Execute(f, w.prog); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}

	// if the program's an executable then compile it and remove the intermediary source
	if w.prog.IsExecutable {
		outfile := filepath.Join(w.tempdir, w.Config.OutputName)
		stderr := strings.Builder{}

		cmd := exec.Command("go", "build", "-o", outfile)
		cmd.Dir = w.tempdir
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to compile %q: %v\n%s\n", srcfile, err, stderr.String())
		}
		if err := os.Remove(srcfile); err != nil {
			return err
		}
	}
	return nil
}

func (w *write) copyAssets() error {
	type filecopy struct{ to, from string }
	files := []filecopy{{
		to:   filepath.Join(w.cssdir, "main.css"),
		from: filepath.Join(w.assetsdir, "main.css"),
	}, {
		to:   filepath.Join(w.jsdir, "main.js"),
		from: filepath.Join(w.assetsdir, "main.js"),
	}}

	for _, f := range files {
		from, err := os.Open(f.from)
		if err != nil {
			return err
		}
		defer from.Close()

		to, err := os.OpenFile(f.to, os.O_CREATE|os.O_WRONLY, w.fperm)
		if err != nil {
			return err
		}
		defer to.Close()

		// copy and sync to disk
		if _, err = io.Copy(to, from); err != nil {
			return err
		}
		if err := to.Sync(); err != nil {
			return err
		}
	}
	return nil
}