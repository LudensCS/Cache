package consistenthash

import (
	"hash/crc32"
	"slices"
	"sort"
	"strconv"
)

type Hash func([]byte) uint32

// 存储所有的hash keys
type Map struct {
	hash     Hash              //Hash函数
	replicas int               //虚拟节点数
	keys     []uint32          //sorted
	hashMap  map[uint32]string //虚拟节点与真实节点的映射表
}

// 创建Map实例
func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		keys:     make([]uint32, 0),
		hashMap:  make(map[uint32]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 添加真实节点的方法
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := range m.replicas {
			hash := m.hash([]byte(strconv.FormatInt(int64(i), 10) + key))
			m.hashMap[hash] = key
			m.keys = append(m.keys, hash)
		}
	}
	slices.Sort(m.keys)
}

// 得到输入key值对应的真实节点名称
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := m.hash([]byte(key))
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
