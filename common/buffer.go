package common

type buffer struct {
	bytes    []byte
	writeIdx int
	readIdx  int
}
