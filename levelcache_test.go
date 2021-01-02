package levelcache

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	cache, _ := NewCache("/tmp/levelcache_test", 5*time.Second)

	key := []byte("key1")
	value := []byte("value1")

	cache.Set(key, value, 3)

	v, _ := cache.Get(key)
	sv := string(v)
	t.Log(sv)
	if sv != "value1" {
		t.Error("get value error")
	}

	time.Sleep(6 * time.Second)

	v, err := cache.Get(key)
	t.Log(string(v), err)
}
