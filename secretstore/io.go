package secretstore

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

func readFull(in io.Reader, ba []byte) (err error) {
	_, err = io.ReadFull(in, ba)
	return
}

func read[T any](in io.Reader) (v T, err error) {
	err = binary.Read(in, binary.BigEndian, &v)
	return
}

var readSize = read[uint16]

func randRead(ba []byte) (err error) {
	err = readFull(rand.Reader, ba)
	if err != nil {
		err = fmt.Errorf("failed to read random bytes: %w", err)
		return
	}

	return
}
