package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	//protects m
	mu sync.Mutex
	m  map[string]*call
}

// 针对相同key，无论 Do 被调用多少次，函数 fn 都只会被调用一次(fn集成在geecache中就是节点选择函数)
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	//延迟初始化，提高内存使用效率
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	//若相同key请求正在进行中，
	if c, ok := g.m[key]; ok {
		//读完 m 后，解锁
		g.mu.Unlock()
		//阻塞等待
		c.wg.Wait()
		return c.val, c.err
	}
	//若key为新请求
	c := new(call)
	//发起新请求前加锁
	c.wg.Add(1)
	// 添加到 g.m，表明 key 已经有对应的请求在处理
	g.m[key] = c
	//添加完 key 后，解锁
	g.mu.Unlock()

	//c为指针，相应 key 对应的 c 值也会改变
	c.val, c.err = fn()
	//请求结束
	c.wg.Done()

	g.mu.Lock()
	//更新 g.m
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
