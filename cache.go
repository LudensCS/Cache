// 并发控制
package Cache

import (
	"sync"

	"github.com/LudensCS/Cache/lru"
)

// 多线程安全缓存
type cache struct {
	mutex      sync.Mutex
	lru        *lru.Cache
	CacheBytes int64
}

func (c *cache) Add(key string, value ByteView) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.CacheBytes, nil)
	}
	c.lru.Add(key, value)
}
func (c *cache) Get(key string) (value ByteView, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil {
		return ByteView{}, false
	}
	if value, ok := c.lru.Get(key); ok {
		return value.(ByteView), true
	}
	return ByteView{}, false
}
