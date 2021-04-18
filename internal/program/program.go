package program

type Program struct {
	// The name of the Go package.
	PkgName      string
	ListenAddr   string
	IsExecutable bool
	Handlers     []Handler
}

type Handler struct {
	Name string
	Path string
	File string
}
