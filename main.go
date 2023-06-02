package main

import (
	"flag"
	"fmt"
	"geecache"
	"log"
	"net/http"
)

// 数据源
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// 创建缓存
func createGroup() *geecache.Group {
	return geecache.NewGroup("score", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// 启动缓存服务器
func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	//本机节点的HTTPPool对象（节点选择）
	peers := geecache.NewHTTPPool(addr)
	//为每个传入的节点添加对应的httpGetter
	peers.Set(addrs...)
	//将节点选择器集成到缓存中
	gee.RegisterPeers(peers)
	log.Println("geecache is running at ", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// 本节点（本机）启动API服务与用户交互
func startAPIServer(apiAddr string, gee *geecache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			//Get方法专门用来处理只有一个参数的参数值，当参数存在多个值时返回第一个值
			key := r.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			//设置HTTP头 Content-Type表示发送响应的数据类型描述
			w.Header().Set("Conntent-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		}))
	log.Println("fontend server is running at", apiAddr)
	//监听 apiAddr（/localhost:9999）端口
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	//试用flag包，解析命令行参数
	//绑定
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	//解析
	flag.Parse()

	//本节点（本机）开启节点服务的地址和端口
	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	// 将addrMap中的value放入切片中
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	//创建一个缓存空间Group,返回*geecache.Group
	gee := createGroup()
	//若api为true，开启api服务，用户通过端口9999访问
	if api {
		go startAPIServer(apiAddr, gee)
	}
	//启动缓存服务器
	//addrs的值是["http://localhost:8001","http://localhost:8002","http://localhost:8003"]
	startCacheServer(addrMap[port], addrs, gee)
}
