package levelcache

import (
	"bytes"
	"encoding/binary"
)

type CacheBin struct {
	AddTime   int64
	Expire    int64
	ValueSize int64
}

type CacheValue struct {
	*CacheBin
	Value []byte
}

func parseBinary(bs []byte) (*CacheValue, error) {
	buf := bytes.NewBuffer(bs)

	cv := &CacheValue{}
	cv.CacheBin = &CacheBin{}

	if err := binary.Read(buf, binary.LittleEndian, cv.CacheBin); err != nil {
		return nil, err
	}

	cv.Value = make([]byte, cv.ValueSize)
	if err := binary.Read(buf, binary.LittleEndian, cv.Value); err != nil {
		return nil, err
	}

	return cv, nil
}

func (cv *CacheValue) toBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.LittleEndian, cv.CacheBin); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.LittleEndian, cv.Value); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
