package localdb

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"

	"github.com/cockroachdb/pebble"
)

var (
	ErrNotFound = pebble.ErrNotFound

	dbPath = "/tmp/localdb"
)

func init() {
	flag.StringVar(&dbPath, "localdb", dbPath, "local DB path")
}

type DB[T any] struct {
	raw *pebble.DB
}

func Exists(bucket string) (exists bool, err error) {
	stat, err := os.Stat(filepath.Join(dbPath, bucket))
	if os.IsNotExist(err) {
		return
	} else if err != nil {
		return
	}

	exists = stat.IsDir()
	return
}

func Open[T any](bucket string) (db DB[T], err error) {
	db.raw, err = pebble.Open(filepath.Join(dbPath, bucket), nil)
	return
}

func (db DB[T]) Close() (err error) {
	return db.raw.Close()
}

func (db DB[T]) Has(key []byte) (ok bool, err error) {
	err = db.GetRaw(key, func(data []byte) error { return nil })
	if err == nil {
		ok = true
	} else if err == ErrNotFound {
		err = nil
	}
	return
}

func (db DB[T]) Get(key []byte) (v T, err error) {
	err = db.GetRaw(key, func(data []byte) error {
		return json.Unmarshal(data, &v)
	})
	return
}

func (db DB[T]) GetRaw(key []byte, processData func(data []byte) error) (err error) {
	data, closer, err := db.raw.Get(key)
	if err != nil {
		return
	}

	defer closer.Close()
	err = processData(data)
	return
}

func (db DB[T]) Set(key []byte, v T) (err error) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}

	return db.SetRaw(key, data)
}

func (db DB[T]) SetRaw(key []byte, data []byte) (err error) {
	batch := db.raw.NewBatch()

	if err = batch.Set(key, data, nil); err != nil {
		return
	}
	if err = batch.Commit(nil); err != nil {
		return
	}

	return
}
