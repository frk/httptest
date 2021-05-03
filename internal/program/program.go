package program

import (
	"sort"
)

type Program struct {
	// The name of the Go package.
	PkgName      string
	RootPath     string
	ListenAddr   string
	IsExecutable bool
	ValidPaths   map[string]string
	SnippetTypes []string
	Users        map[string]string
	SessionName  string
}

func (p Program) Imports() []string {
	imports := []string{
		"html/template",
		"log",
		"net/http",
		"path/filepath",
		"strings",
	}
	if p.IsExecutable {
		imports = append(imports, "os")
	}
	if len(p.Users) > 0 {
		imports = append(imports, []string{
			"time",
			"sync",
			"golang.org/x/crypto/bcrypt",
			"crypto/rand",
			"encoding/base64",
			"context",
		}...)
	}

	sort.Strings(imports)
	return imports
}
