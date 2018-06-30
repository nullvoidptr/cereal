package cereal

import (
	"bytes"
	"strconv"
)

// buffer is a wrapper to bytes.buffer with some additional methods
// to simplify some common use cases
type buffer struct {
	*bytes.Buffer
}

// newBuffer returns a new Buffer
func newBuffer(data []byte) *buffer {
	return &buffer{Buffer: bytes.NewBuffer(data)}
}

// readLineStr reads the next line from Buffer and returns as a string
// Note: The returned string will have the newline character removed.
// If EOF occurs prior to the next newline, the full text until EOF
// will be returned and the io.EOF error will be returned as well.
func (b *buffer) readLineStr() (string, error) {

	line, err := b.ReadBytes('\n')
	if err != nil {
		return string(line), err
	}

	// Trim the newline from buffer before returning
	return string(line[:len(line)-1]), nil
}

// readLineInt is identical to ReadLineStr but performs a conversion
// of the string into an int.
func (b *buffer) readLineInt() (int, error) {

	line, err := b.readLineStr()
	if err != nil {
		return 0, err
	}

	val, err := strconv.Atoi(line)
	if err != nil {
		return 0, err
	}

	return val, nil
}
