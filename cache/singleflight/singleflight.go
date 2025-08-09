package singleflight

import (
	"sync"
)

// 进行中或已经结束的请求
type call struct {
	wg  sync.WaitGroup
	val any
	err error
}

// 管理不同key的请求
type Group struct {
	mutex sync.Mutex
	mp    map[string]*call
}

func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mutex.Lock()
	if g.mp == nil {
		g.mp = make(map[string]*call)
	}
	//已有key值相同的请求正在处理,进行等待
	if c, ok := g.mp[key]; ok {
		g.mutex.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.mp[key] = c //登记到map,表明key值已有请求正在进行
	g.mutex.Unlock()

	c.val, c.err = fn() //处理请求
	c.wg.Done()
	g.mutex.Lock()
	delete(g.mp, key) //处理完毕,删除记录
	g.mutex.Unlock()

	return c.val, c.err
}
