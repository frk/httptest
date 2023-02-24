package types

type DataList[T any] struct {
	// The total number of available data.
	Total int
	// The list of retrieved data.
	Data []T
}
