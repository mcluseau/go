package secretstore

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"syscall"

	"golang.org/x/crypto/argon2"
)

type Store struct {
	unlocked bool
	key      [32]byte
	salt     [aes.BlockSize]byte
	keys     []keyEntry
}

type keyEntry struct {
	hash   [64]byte
	encKey [32]byte
}

func New() (s *Store) {
	s = &Store{}
	syscall.Mlock(s.key[:])
	return
}

func Open(path string) (s *Store, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}

	defer f.Close()

	s = New()
	_, err = s.ReadFrom(bufio.NewReader(f))

	return
}

func (s *Store) SaveTo(path string) (err error) {
	f, err := os.OpenFile(path, syscall.O_CREAT|syscall.O_TRUNC|syscall.O_WRONLY, 0600)
	if err != nil {
		return
	}

	defer f.Close()

	out := bufio.NewWriter(f)

	_, err = s.WriteTo(out)
	if err != nil {
		return
	}

	err = out.Flush()
	if err != nil {
		return
	}

	return
}

func (s *Store) Close() {
	memzero(s.key[:])
	syscall.Munlock(s.key[:])
	s.unlocked = false
}

func (s *Store) IsNew() bool {
	return len(s.keys) == 0
}

func (s *Store) Unlocked() bool {
	return s.unlocked
}

func (s *Store) Init(passphrase []byte) (err error) {
	err = randRead(s.key[:])
	if err != nil {
		return
	}
	err = randRead(s.salt[:])
	if err != nil {
		return
	}

	s.AddKey(passphrase)

	s.unlocked = true

	return
}

func (s *Store) ReadFrom(in io.Reader) (n int64, err error) {
	memzero(s.key[:])
	s.unlocked = false

	defer func() {
		if err != nil {
			log.Output(2, fmt.Sprintf("failed after %d bytes", n))
		}
	}()

	readFull := func(ba []byte) {
		var nr int
		nr, err = io.ReadFull(in, ba)
		n += int64(nr)
	}

	// read the salt
	readFull(s.salt[:])
	if err != nil {
		return
	}

	// read the (encrypted) keys
	s.keys = make([]keyEntry, 0)
	for {
		k := keyEntry{}
		readFull(k.hash[:])
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		readFull(k.encKey[:])
		if err != nil {
			return
		}

		s.keys = append(s.keys, k)
	}
}

func (s *Store) WriteTo(out io.Writer) (n int64, err error) {
	write := func(ba []byte) {
		var nr int
		nr, err = out.Write(ba)
		n += int64(nr)
	}

	write(s.salt[:])
	if err != nil {
		return
	}

	for _, k := range s.keys {
		write(k.hash[:])
		if err != nil {
			return
		}

		write(k.encKey[:])
		if err != nil {
			return
		}
	}

	return
}

var ErrNoSuchKey = errors.New("no such key")

func (s *Store) Unlock(passphrase []byte) (ok bool) {
	key, hash := s.keyPairFromPassword(passphrase)
	memzero(passphrase)
	defer memzero(key[:])

	var idx = -1
	for i := range s.keys {
		if hash == s.keys[i].hash {
			idx = i
			break
		}
	}

	if idx == -1 {
		return
	}

	s.decryptTo(s.key[:], s.keys[idx].encKey[:], &key)

	s.unlocked = true
	return true
}

func (s *Store) AddKey(passphrase []byte) {
	key, hash := s.keyPairFromPassword(passphrase)
	memzero(passphrase)

	defer memzero(key[:])

	k := keyEntry{hash: hash}

	encKey := s.encrypt(s.key[:], &key)
	copy(k.encKey[:], encKey)

	s.keys = append(s.keys, k)
}

func (s *Store) keyPairFromPassword(password []byte) (key [32]byte, hash [64]byte) {
	keySlice := argon2.IDKey(password, s.salt[:], 1, 64*1024, 4, 32)

	copy(key[:], keySlice)
	memzero(keySlice)

	hash = sha512.Sum512(key[:])

	return
}

func (s *Store) NewEncrypter(iv [aes.BlockSize]byte) cipher.Stream {
	if !s.unlocked {
		panic("not unlocked")
	}
	return newEncrypter(iv, &s.key)
}

func (s *Store) NewDecrypter(iv [aes.BlockSize]byte) cipher.Stream {
	if !s.unlocked {
		panic("not unlocked")
	}
	return newDecrypter(iv, &s.key)
}

func (s *Store) encrypt(src []byte, key *[32]byte) (dst []byte) {
	dst = make([]byte, len(src))
	newEncrypter(s.salt, key).XORKeyStream(dst, src)
	return
}

func (s *Store) decryptTo(dst []byte, src []byte, key *[32]byte) {
	newDecrypter(s.salt, key).XORKeyStream(dst, src)
}

func newEncrypter(iv [aes.BlockSize]byte, key *[32]byte) cipher.Stream {
	c, err := aes.NewCipher(key[:])
	if err != nil {
		panic(fmt.Errorf("failed to init AES: %w", err))
	}

	return cipher.NewCFBEncrypter(c, iv[:])
}

func newDecrypter(iv [aes.BlockSize]byte, key *[32]byte) cipher.Stream {
	c, err := aes.NewCipher(key[:])
	if err != nil {
		panic(fmt.Errorf("failed to init AES: %w", err))
	}

	return cipher.NewCFBDecrypter(c, iv[:])
}
