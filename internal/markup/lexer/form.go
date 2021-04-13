package lexer

type FormToken uint

const (
	_ FormToken = iota
	// TODO
)

func Form(data []byte, class map[FormToken]string) ([]byte, error) {
	// TODO
	return data, nil
}
