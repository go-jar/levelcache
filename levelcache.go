package levelcache

import (
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

type Cache struct {
	db *leveldb.DB

	stopJanitorCh chan bool
}

func NewCache(cacheDir string, janitorInterval time.Duration) (*Cache, error) {
	db, err := leveldb.OpenFile(cacheDir, nil)
	if err != nil {
		return nil, err
	}

	c := &Cache{
		db:            db,
		stopJanitorCh: make(chan bool),
	}

	go c.runJanitor(janitorInterval)

	return c, nil
}

func (c *Cache) runJanitor(janitorInterval time.Duration) {
	ticker := time.NewTicker(janitorInterval)

	for {
		select {
		case <-c.stopJanitorCh:
			return
		case <-ticker.C:
			now := time.Now().Unix()
			iter := c.db.NewIterator(nil, nil)
			for iter.Next() {
				key := iter.Key()
				cv, err := parseBinary(iter.Value())
				if err != nil {
					c.Delete(key)
					continue
				}
				if cv.Expire == 0 {
					continue
				}
				if now-cv.AddTime > cv.Expire {
					c.Delete(key)
				}
			}
			iter.Release()
		}
	}
}

func (c *Cache) Delete(key []byte) error {
	return c.db.Delete(key, nil)
}

func (c *Cache) Free() {
	c.db.Close()

	c.stopJanitorCh <- true
	close(c.stopJanitorCh)
}

func (c *Cache) Set(key, value []byte, expireSeconds int64) error {
	cb := &CacheBin{
		AddTime:   time.Now().Unix(),
		Expire:    expireSeconds,
		ValueSize: int64(len(value)),
	}

	cv := &CacheValue{
		CacheBin: cb,
		Value:    value,
	}

	bv, err := cv.toBinary()
	if err != nil {
		return err
	}

	return c.db.Put(key, bv, nil)
}

func (c *Cache) Get(key []byte) ([]byte, error) {
	bv, err := c.db.Get(key, nil)
	if err != nil {
		if err == errors.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	cv, err := parseBinary(bv)
	if err != nil {
		return nil, err
	}

	if cv.Expire != 0 {
		if time.Now().Unix()-cv.AddTime > cv.Expire {
			c.Delete(key)
			return nil, nil
		}
	}
	return cv.Value, nil
}
