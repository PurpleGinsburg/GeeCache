package lru

import "container/list"

/*
  LRU(Least Recently Used)算法实现
  FIFO(First In First Out)和LFU(Least Frequently Used)的结合
*/

// Value use Len to count how mant bytes it takes
// 返回值所占用内存的大小
type Value interface {
	Len() int
}

// entry表示双向链表节点的数据类型
// 在链表中保存key的优点，淘汰队首节点时，用key从字典中删除对应的映射
type entry struct {
	key string
	//值是实现了Value接口的任意类型
	value Value
}

// 存放map 和 list
type Cache struct {
	//允许使用的最大内存
	maxBytes int64
	//当前已使用的最大内存 包含了双向链表中 元素entry中key和value两者所占内存之和
	nbytes int64
	//双向链表(double linked list)实现的队列
	ll *list.List
	//字典(map)
	cache map[string]*list.Element
	//某条记录被删除时的回调函数
	//optinal and executed when an entry is pruged
	OnEvicted func(key string, value Value)
}

// 初始化
// New is the Constructor of Cache
func NewLruCache(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 增 & 改
// Add adds a value to the cache.
func (c *Cache) Add(key string, value Value) {

	// 改/更新
	//根据key找map中对应的element——ele，ele同时也是链表元素结构体
	if ele, ok := c.cache[key]; ok {
		//移到列表最前
		c.ll.MoveToFront(ele)
		//ele.Value是链表元素内容 any为空指针 .(*entry)是类型断言 kv转换为*entry类型
		kv := ele.Value.(*entry)
		//计算新加值 新值int64(value.Len()) 旧值int64(kv.value.Len()) 语法糖 指针直接指向value
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		//增 Add adds a value to the cache
		//不存在，则新建一个节点
		ele := c.ll.PushFront(&entry{key, value})
		//添加map映射关系
		c.cache[key] = ele
		//key 和 value 所占内存之和
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// 删 RemoveOldest removes the oldest item
func (c *Cache) RemoveOldest() {
	//取列表末尾元素
	ele := c.ll.Back()
	if ele != nil {
		//从双向链表中删除
		c.ll.Remove(ele)
		//类型断言
		kv := ele.Value.(*entry)
		//从map中删去映射关系
		delete(c.cache, kv.key)
		//更新占用内存
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		//有回调函数就调用回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// 查 Get look ups a key's value
func (c *Cache) Get(key string) (value Value, ok bool) {
	//若节点存在，移动到front，并返回找到的值
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// 测试用
// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
