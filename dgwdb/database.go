package dgwdb

import (
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/errors"
	"github.com/btcsuite/goleveldb/leveldb/filter"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	"github.com/inconshreveable/log15"

	"sync"
)

type LDBDatabase struct {
	filename string
	db       *leveldb.DB
	log      log15.Logger
	quitLock sync.Mutex
	quitChan chan chan error
}


func NewLDBDatabase(file string, cache int, handles int) (*LDBDatabase, error) {
	//logger := log.New()

	if cache < 16 {
		cache = 16
	}
	if handles < 16 {
		handles = 16
	}

	db, err := leveldb.OpenFile(file, &opt.Options{
		OpenFilesCacheCapacity: handles,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache / 4 * opt.MiB,
		Filter:                 filter.NewBloomFilter(10),
	})
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(file, nil)
	}

	if err != nil {
		return nil, err
	}

	return &LDBDatabase{
		filename: file,
		db:       db,
	}, nil
}