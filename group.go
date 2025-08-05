package Cache

import (
	"fmt"
	"log"
	"sync"
)

type GetterFunc func(key string) ([]byte, error)

type Getter interface {
	Get(string) ([]byte, error)
}

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// 实例化Gruop对象
func NewGroup(name string, CacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("getter is nil!")
	}
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{CacheBytes: CacheBytes},
	}
	mu.Lock()
	defer mu.Unlock()
	groups[name] = g
	return g
}
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	if g, ok := groups[name]; ok {
		return g
	}
	return nil
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if value, ok := g.mainCache.Get(key); ok {
		log.Println("Cache hit")
		return value, nil
	}
	return g.Load(key)
}
func (g *Group) Load(key string) (ByteView, error) {
	return g.GetLocally(key)
}
func (g *Group) GetLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: CloneBytes(bytes)}
	g.PopulateCache(key, value)
	return value, nil
}
func (g *Group) PopulateCache(key string, value ByteView) {
	g.mainCache.Add(key, value)
}
