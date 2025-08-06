package lru

import (
	"container/list"
)

// LRU cache
type Cache struct {
	maxBytes  int64
	nowBytes  int64
	lst       *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

// 构造函数
func New(maxBytes int64, OnEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		nowBytes:  0,
		lst:       list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: OnEvicted,
	}
}

// 双向链表存储的结点类型
type entry struct {
	key   string
	value Value
}

// 实现了Value接口的值都可被Cache接受
type Value interface {
	Len() int
}

// 从缓存中查询key
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.lst.MoveToBack(ele)
		kv := ele.Value.(*entry) //类型断言,返回(entry,ok),失败则panic
		return kv.value, true
	}
	return nil, false
}

// 缓存淘汰
func (c *Cache) RemoveOldest() {
	ele := c.lst.Front()
	if ele != nil {
		c.lst.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nowBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// 添加或修改缓存键值对
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.lst.MoveToBack(ele)
		kv := ele.Value.(*entry)
		c.nowBytes += int64(value.Len()) - int64(kv.value.Len())
		ele.Value = value
	} else {
		ele := c.lst.PushBack(&entry{key, value})
		c.nowBytes += int64(value.Len()) + int64(len(key))
		c.cache[key] = ele
	}
	for c.maxBytes != 0 && c.nowBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.lst.Len()
}
