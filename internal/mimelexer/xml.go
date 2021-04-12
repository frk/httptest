package mimelexer

type XMLToken uint

const (
	_ XMLToken = iota
	// TODO
)

func XML(data []byte, class map[XMLToken]string) ([]byte, error) {
	// TODO
	return data, nil
}
