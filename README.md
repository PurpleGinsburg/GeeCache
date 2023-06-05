# GeeCache

GeeCache —— 分布式缓存框架，仿groupcache实现，实现了如下功能

1、LRU(Least Recently Used)缓存策略 —— lru.go 文件 ：
container/list 包中的 list.List 实现一双向链表，利用 map 存储该链表中各节点的指针，map和list配合实现增删改查。

2、单机并发缓存 —— ByteView.go 文件、cache.go 文件、geecache.go 文件 ：
ByteView.go 文件 ：负责创建只读数据结构 ByteView 为缓存值，实现 Len() int 方法，接口 Value 的实例化 ；
cache.go 文件 ：创建cache结构体，实例化lru，并利用互斥锁添加并发特性 ；
geecache.go 文件：
V0.1 ：实现缓存不存在时的回调接口型函数Getter，主体结构 Group（包括 Getter 和 cache）。通过Get 和 Load 完成单机并发缓存。

3、一致性哈希 —— consistenthash.go 文件 ：
针对节点数量变化的场景，采用环状结构（利用[]int 实现，用取余数的方式通过 hashMap 映射得到真实的节点）；针对针对数据倾斜问题，采用虚拟节点。哈希算法采取依赖注入的方式，为方便测试默认为 crc32.ChecksumIEEE 算法。

5、分布式节点 —— peer.go 文件 ：
peer.go 文件 ：抽象出 2 个接口 PeerPicker 和 PeerGetter 。接口 PeerPicker 的 PickPeer() 方法用于根据传入的 key 选择相应节点 PeerGetter ；接口 PeerGetter 的 Get() 方法用于从对应 group 查找缓存值， PeerGetter 即 功能4 将要实现的 客户端。

4、节点间 HTTP 通信 —— http.go 文件 、geecache.go 文件：
http.go 文件 ：结构体 HTTPPool 作为承载节点间 HTTP 通信的核心数据结构，该结构体中 basePath 作为节点间通讯地址的前缀，约定访问路径格式为 /<basepath>/<groupname>/<key> 。
服务端 ： ServeHTTP方法 —— 实现逻辑是先判断访问路径的前缀是否是 basePath（strings.HasPrefix(r.URL.Path, p.basePath)），不是 —— 返回错误，是 —— 通过 groupname 得到 group 实例，再使用 group.Get(key) 获取缓存数据，最终使用 w.Write() 将缓存值作为响应体返回。
客户端类 ：结构体 httpGetter，实例化 PeerGetter 接口，利用 http.Get() 方式获取返回值 ；每个远程节点对应一个 httpGetter。
HTTPPool 结构体通过 ServeHTTP 方法 响应其他节点需求；通过成员变量 peers 根据具体的 key 选择节点（实现接口 PeerPicker），通过 httpGetters（map类型）映射远程节点与对应的 httpGetter（实现接口 PeerGetter）。
geecache.go 文件：
V0.2 ：在Group中添加 PeerPicker 集成上述功能。
  
5、防止缓存击穿 —— singleflight.go 文件 ：
防止同时访问数据库缓存造成缓存击穿和穿透，针对可能同时并发的相同请求，通过同步锁 sync.WaitGroup 锁避免重入，将原来的 load 的逻辑，使用 g.loader.Do 包裹起来。
