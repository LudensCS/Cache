package cache

import "slices"

// 只读数据结构,表示缓存值
type ByteView struct {
	//存储真实缓存值
	b []byte
}

// 实现Value接口
func (View ByteView) Len() int {
	return len(View.b)
}

// ByteView是只读的,使用该方法返回一个拷贝,防止缓存值被外部程序修改
func (View ByteView) ByteSlice() []byte {
	return CloneBytes(View.b)
}
func (View ByteView) String() string {
	return string(View.b)
}
func CloneBytes(b []byte) []byte {
	return slices.Clone(b)
}
