package secretstore

import (
	"crypto/aes"
	"crypto/cipher"
	"io"
)

func (s *Store) NewReader(reader io.Reader) (r io.Reader, err error) {
	iv := [aes.BlockSize]byte{}

	err = readFull(reader, iv[:])
	if err != nil {
		return
	}

	r = storeReader{reader, s.NewDecrypter(iv)}
	return
}

type storeReader struct {
	reader    io.Reader
	decrypter cipher.Stream
}

func (r storeReader) Read(ba []byte) (n int, err error) {
	n, err = r.reader.Read(ba)

	if n > 0 {
		r.decrypter.XORKeyStream(ba[:n], ba[:n])
	}

	return
}

func (s *Store) NewWriter(writer io.Writer) (r io.Writer, err error) {
	iv := [aes.BlockSize]byte{}

	if err = randRead(iv[:]); err != nil {
		return
	}

	_, err = writer.Write(iv[:])
	if err != nil {
		return
	}

	r = storeWriter{writer, s.NewEncrypter(iv)}
	return
}

type storeWriter struct {
	writer    io.Writer
	encrypter cipher.Stream
}

func (r storeWriter) Write(ba []byte) (n int, err error) {
	if len(ba) == 0 {
		return
	}

	encBA := make([]byte, len(ba))
	r.encrypter.XORKeyStream(encBA, ba)

	n, err = r.writer.Write(encBA)

	return
}
