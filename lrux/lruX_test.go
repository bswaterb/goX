package lrux

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLRU(t *testing.T) {
	cache := NewLRUCache(2, 2)
	cache.Set("a", 1)
	cache.Set("b", 2)

	_, found := cache.Get("a")
	assert.True(t, found)
	assert.False(t, cache.eleInOld("a"))

	time.Sleep(3 * time.Second)

	// 此时再次访问 key 为 "a" 的数据，内部会将该 key 从 young 转移到 old
	cache.Get("a")
	assert.True(t, cache.eleInOld("a"))

	cache.Set("c", 3)
	cache.Set("d", 4)
	cache.Set("e", 5)

	_, found = cache.Get("b")
	// 此时 b 应当已经被淘汰
	assert.False(t, found)

	_, found = cache.Get("a")
	assert.True(t, found)
	assert.True(t, cache.eleInOld("a"))

}
