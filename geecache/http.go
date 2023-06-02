package geecache

import (
	"fmt"
	"geecache/consistenthash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

/*
    分布式缓存实现节点间通信，建立基于HTTP的通信机制
	一个节点启动了HTTP服务端就可以被其他节点返回
	也有其他方式，如RPC等
*/

//结构体HTTPPool作为承载节点间HTTP通信的核心数据结构（包括客户端和服务端）

// 加前缀(BasePath)用于节点间访问，因为主机上还可能承载其他服务
// 约定访问格式为/<basepath>/<groupname>/<key>
// 通过groupname得到group实例，再使用group.Get(key)获取缓存数据
const (
	defaultBasePath = "/_geecache/"
	//默认虚拟节点值
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	//this peer'base URL, e.g. "http:////example.net:8000"
	//记录自己的地址，包括主机名/IP和端口
	self string
	//作为节点间通讯地址的前缀
	//默认是defaultBasePath，"http://example.com/_geecache/"开头的请求，用于节点间访问
	basePath string

	/*
	   添加节点选择功能
	*/
	mu sync.Mutex //guards peers and httpGetters
	//一致性哈希表
	peersMap *consistenthash.Map
	//存放 http客户端类 ,映射远程节点与对应的httpGerrer,通过httpGerrer找到远程地址的base URL
	httpGetters map[string]*httpGetter //keyed by e.g. "http://10.0.0.2:8008"
}

// NewHTTPPool initializes an HTTP pool of peers
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		//记录自己地址，包括主机名/IP和端口
		self: self,
		//作为节点间通讯地址的前缀
		basePath: defaultBasePath,
		//实例化一致性哈希
		peersMap: consistenthash.NewMap(defaultReplicas, nil),
		//初始化
		httpGetters: make(map[string]*httpGetter),
	}
}

// log info with server name
func (hp *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", hp.self, fmt.Sprintf(format, v...))
}

// ServerHTTP handle all http requests(http标准库做法) 使HTTPPool实现handler http.Handler接口
func (hp *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//判读请求的前缀是否为p.basePath
	if !strings.HasPrefix(r.URL.Path, hp.basePath) {
		panic("HTTPPool serving unexpected path:" + r.URL.Path)
	}
	hp.Log("%s %s", r.Method, r.URL.Path)
	// r.URL.Path ———— /<basepath>/<groupname>/<key> required
	//r.URL.Path[len(hp.basePath):] ———— /<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(hp.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

/*
    实现 PeerPicker 接口
	根据具体的key,创建http客户端从远程节点获取缓存值
*/

// 添加传入的节点,并为每个节点创建了一个httpGetter
// Set updates the pool'list of peers.
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	//p.peers = consistenthash.NewMap(defaultReplicas,nil)
	p.peersMap.Add(peers...)
	//p.httpGetters =make(map[string]*httpGetter,len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// 包装一致性哈希的Get方法，根据具体key，选择节点，返回节点对应的客户端
// PickPeer picks a peer according to key
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	//选择节点，若选择的节点不存在或选出的节点为本机节点则返回false
	if peer := p.peersMap.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

// 检查 HTTPPool 是否实现了接口 PeerPicker ，若没有会编译出错
var _ PeerPicker = (*HTTPPool)(nil)

/*
   实现 PeerGetter 接口
*/

// http客户端类，实现PeerGetter接口
type httpGetter struct {
	baseURL string
}

// Get HTTP客户端httpGetter提交的get请求，访问远程节点
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	//字符串拼接
	u := fmt.Sprintf(
		"%v%v%v",
		//表示将要访问的远程节点的地址
		h.baseURL,
		//QueryEscape函数对s进行转码使之可以安全的用在URL查询里
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	//发起http get请求
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	//关闭挂起的Body(io.ReadCloser)
	defer res.Body.Close()

	//检查状态码
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server return : %v", res.Status)
	}

	//读取消息体相应内容
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body :%v", err)
	}

	return bytes, nil
}

// 检查 httpGetter 是否实现了接口 PeerGetter ，若没有会编译出错
var _ PeerGetter = (*httpGetter)(nil)
