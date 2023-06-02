package geecache

import (
	"geecache/lru"
	"sync"
)

/*
    并发控制
    为lru.Cache添加并发特性
	实例化lru，封装get和add方法
*/

type cache struct {
	//互斥锁 不用读写锁因为都涉及写操作（LRU将最近访问元素移至链表头）
	//TODO 有优化空间
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

// 延迟初始化(Lazy Initialization)方法：一个对象的延迟初始化意味着该对象的创建将会延迟至第一次使用该对象时。主要用于提高性能，并减少程序内存要求
// 若c.lru为空再创建实例
func (c *cache) add(key string, value ByteView) {
	//写锁
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		c.lru = lru.NewLruCache(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	//写锁 （lru.Get会将最近访问元素添加至链表头，因此不用RLock）
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}
