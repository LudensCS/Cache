package bloomfilter

import (
	"hash/maphash"
	"math"

	"github.com/bits-and-blooms/bitset"
)

// 默认误判率
const defaultP = 1e-4

// 布隆过滤器
// 存在P的概率将不存在数据误判为存在
type Bloomfilter struct {
	hashSeeds []maphash.Seed
	bitmap    *bitset.BitSet
	mod       int
}

// 初始化布隆过滤器,数据条数为n
func New(n int) *Bloomfilter {
	k := int(math.Ceil(-math.Log(defaultP) / math.Ln2))
	m := int(math.Ceil(float64(n*k) / math.Ln2))
	BF := &Bloomfilter{
		hashSeeds: make([]maphash.Seed, 0),
		mod:       m,
		bitmap:    bitset.New(uint(m)),
	}
	for range k {
		BF.hashSeeds = append(BF.hashSeeds, maphash.MakeSeed())
	}
	return BF
}

// 获取对应哈希值
func (BF *Bloomfilter) GetHash(key string, seed maphash.Seed) uint64 {
	var f maphash.Hash
	f.SetSeed(seed)
	f.WriteString(key)
	return f.Sum64()
}

// 增加数据key到布隆过滤器中
func (BF *Bloomfilter) Add(key string) {
	for _, seed := range BF.hashSeeds {
		val := BF.GetHash(key, seed)
		val %= uint64(BF.mod)
		BF.bitmap.Set(uint(val))
	}
}

// 查询key值是否可能存在
func (BF *Bloomfilter) Query(key string) bool {
	for _, seed := range BF.hashSeeds {
		val := BF.GetHash(key, seed)
		val %= uint64(BF.mod)
		if !BF.bitmap.Test(uint(val)) {
			return false
		}
	}
	return true
}
