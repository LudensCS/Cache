package cache

import (
	"fmt"
	"log"
	"sync"

	"github.com/LudensCS/Cache/cache/cachepb"
	"github.com/LudensCS/Cache/cache/singleflight"
)

type GetterFunc func(key string) ([]byte, error)

type Getter interface {
	Get(string) ([]byte, error)
}

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// 缓存组,一个缓存组可以有多个分布式节点
// 同名缓存组共享逻辑地址,如果两个节点在同名缓存组内,它们属于同一个子系统
type Group struct {
	name      string
	getter    Getter //回调函数
	mainCache cache
	peers     PeerPicker
	loader    *singleflight.Group //利用singleflight保证同一时间每种请求只会访问数据库一次
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
		loader:    &singleflight.Group{},
	}
	mu.Lock()
	defer mu.Unlock()
	groups[name] = g
	return g
}

// 通过name寻找对应的Group实例
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	if g, ok := groups[name]; ok {
		return g
	}
	return nil
}

// 注册一个peers以选择远端节点
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("group's peer called more than once")
	}
	g.peers = peers
}

// 查询key对应的value
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

// 尝试从远端节点获取缓存,失败则调用GetLocally方法,利用singleflight防止缓存击穿
func (g *Group) Load(key string) (ByteView, error) {
	value, err := g.loader.Do(key, func() (value any, err error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.GetFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[Cache] failed to get from peer :", err)
			}
		}
		return g.GetLocally(key)
	})
	if err != nil {
		return ByteView{}, err
	}
	return value.(ByteView), nil
}

// 从远端节点获取缓存
func (g *Group) GetFromPeer(peer PeerGetter, key string) (ByteView, error) {
	Req := &cachepb.Request{Group: g.name, Key: key}
	Resp := &cachepb.Response{}
	err := peer.Get(Req, Resp)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: CloneBytes(Resp.GetValue())}, nil
}

// 使用回调函数从本地数据源获取key对应的value值并加载到缓存
func (g *Group) GetLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: CloneBytes(bytes)}
	g.PopulateCache(key, value)
	return value, nil
}

// 将key-value加载到缓存
func (g *Group) PopulateCache(key string, value ByteView) {
	g.mainCache.Add(key, value)
}
