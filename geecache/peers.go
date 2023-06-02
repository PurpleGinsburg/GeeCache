package geecache

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
// PeerPicker接口目的是查找对应节点PeerGetter，由HTTPPool实现
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter接口目的是从对应 group 查找缓存值，由httpGetter（http客户端）实现
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
