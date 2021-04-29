package program

type Program struct {
	// The name of the Go package.
	PkgName      string
	RootPath     string
	ListenAddr   string
	IsExecutable bool
	ValidPaths   map[string]string
	SnippetTypes []string
}

func (p Program) Imports() []string {
	if p.IsExecutable {
		return []string{
			"html/template",
			"log",
			"net/http",
			"os",
			"path/filepath",
			"strings",
		}
	}
	return []string{
		"html/template",
		"log",
		"net/http",
		"path/filepath",
		"strings",
	}
}
