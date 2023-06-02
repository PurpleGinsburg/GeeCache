package geecache

import (
	"fmt"
	"geecache/singleflight"
	"log"
	"sync"
)

/*
   缓存不存在时，用户调用回调函数获得源数据
   负责和外部交互，控制缓存存储和获取的主流程
*/

/*
   核心数据结构Group
*/

// A Group is a cache namespace and associated data losded spread over
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	//将查找对应节点PeerGetter功能集成在主流程中
	peerspicker PeerPicker
	//use singleflight.Group to make sure that each key is only fetched once
	loader *singleflight.Group
}

// 全局变量
var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}

	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup returns the names group previously created with NewGroup,or nil if there's no such group
func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()

	if g, ok := groups[name]; ok {
		return g
	} else {
		return nil
	}
}

/*
Get方法————实现返回缓存值 | 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值
*/
func (g *Group) Get(key string) (ByteView, error) {
	//key为空 返回空缓存值，错误
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	//key存在对应缓存值，则返回该缓存值
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	//load调用“回调函数”
	return g.load(key)
}

// 将源数据库添加到缓存中
func (g *Group) populaCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// 调用用户回调函数
func (g *Group) getlocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	//key为空 返回空缓存值，错误
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)} //bytes为切片，利用cloneBytes(bytes)实现深拷贝
	g.populaCache(key, value)
	return value, nil
}

// 从远程节点中获取数据
// 使用实现了 PeerGetter 接口的 httpGetter 从访问远程节点，获取缓存值
func (g *Group) getFromPeer(peergetter PeerGetter, key string) (ByteView, error) {
	bytes, err := peergetter.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

// 缓存不存在 调用回调函数
func (g *Group) load(key string) (value ByteView, err error) {
	//each key is only fetched once(either locally or remotely)
	//regardless of the number of concurrent callers.
	//将原来的 load 的逻辑，用 g.loader.Do 包裹起来
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		//分布式场景下会调用 getFromPeer 从其他节点获取
		//若有peerspicker，则开始选择节点
		if g.peerspicker != nil {
			//若选中远程节点（非本机节点）
			if peergetter, ok := g.peerspicker.PickPeer(key); ok {
				//从远程节点中获取数据
				if value, err := g.getFromPeer(peergetter, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		//选中本机节点或获取失败
		return g.getlocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}

	return
}

/*
   查找对应节点 PeerPicker 模块
*/

// 将实现了 PeerPicker 的接口 HTTPPool 注入到 Group 中
// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peerspicker PeerPicker) {
	if g.peerspicker != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peerspicker = peerspicker
}

//
