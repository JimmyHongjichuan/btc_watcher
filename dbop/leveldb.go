package dbop

import (
	"github.com/btcsuite/goleveldb/leveldb"
	"github.com/btcsuite/goleveldb/leveldb/errors"
	"github.com/btcsuite/goleveldb/leveldb/filter"
	"github.com/btcsuite/goleveldb/leveldb/iterator"
	"github.com/btcsuite/goleveldb/leveldb/opt"
	"github.com/btcsuite/goleveldb/leveldb/util"
	"github.com/inconshreveable/log15"

	"sync"
)

//LDBDatabase leveldb操作类
type LDBDatabase struct {
	filename string
	db       *leveldb.DB
	log      log15.Logger
	quitLock sync.Mutex
	quitChan chan chan error
}

//NewLDBDatabase 新建一个LEVELDB实例
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

//NewIteratorWithPrefix 根据前缀返回iter
func (db *LDBDatabase) NewIteratorWithPrefix(prefix []byte) iterator.Iterator {
	return db.db.NewIterator(util.BytesPrefix(prefix), nil)
}

//Get 查询KEY的VALUE
func (db *LDBDatabase) Get(key []byte) ([]byte, error) {
	data, err := db.db.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return data, nil
}
