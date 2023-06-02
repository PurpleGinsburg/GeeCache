package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

/*
    普通哈希算法的改良版，哈希计算方法不变，通过构建环状的Hash空间代替普通的线性Hash空间
    可用于分布式集群缓存的负载均衡实现
	用环状Hash结构解决普通哈希的扩展性问题（新加节点）及容错性问题（节点下线）
	同时采用虚拟节点对一致性哈希优化，解决数据倾斜和节点雪崩问题
*/

// Hash maps bytes to uint32
// 允许自定义哈希计算方法
type Hash func(data []byte) uint32

// Map constains all hashed keys
type Map struct {
	//Hash计算方法
	hash Hash
	//每个节点的虚拟节点值
	replicas int
	//环状结构
	keys []int //Sorted
	//虚拟节点到真实节点的映射
	hashMap map[int]string
}

// NewEntity creates a Map instance
// 构造函数 允许自定义虚拟节点和Hash算法
func NewMap(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		//默认hash计算方法
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 添加真实节点/机器
// Add adds some keys to the hash.
func (m *Map) Add(keys ...string) {
	//添加0 / 多个正式节点
	for _, key := range keys {
		//给每个节点添加m.replicas个虚拟节点
		for i := 0; i < m.replicas; i++ {
			//虚拟节点名称为strconv.Itoa(i)+key
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			//添加到环上
			m.keys = append(m.keys, hash)
			//虚拟节点和真实节点的映射关系
			m.hashMap[hash] = key
		}
	}
	//虚拟节点添加完毕后排序(升序)
	sort.Ints(m.keys)
}

// 节点选择
// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		//还没有真实节点添加
		return ""
	}

	//计算key的哈希值
	hash := int(m.hash([]byte(key)))
	//Binary search for appropriate replica.
	//二分查找——从[0, n)中取出一个值index，index为[0, n)中最小的使函数f(index)为True的值，且f(index+1)也为True
	//如果无法找到该index值，则该方法为返回n
	//该方法一般用于从一个已经排序的数组中找到某个值所对应的索引
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	//通过hashMap映射得到真实hash值
	//若返回len(m.keys)值，则说明要选择m.keys[0]节点（环状结构），所以采用取余的方式映射真实节点
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
