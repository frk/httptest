package program

type Program struct {
	// The name of the Go package.
	PkgName      string
	RootPath     string
	ListenAddr   string
	IsExecutable bool
	IndexHandler Handler
	Handlers     []Handler
}

type Handler struct {
	Name string
	Path string
	File string
}
