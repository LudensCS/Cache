package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/LudensCS/Cache/cache"
)

var db = map[string]string{
	"jack": "663",
	"Tom":  "78515",
	"lucy": "125",
}

// 创立缓存组
func CreateGroup() *cache.Group {
	return cache.NewGroup("scores", 2<<10, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		},
	))
}

// 启动缓存服务
func StartCacheServer(addr string, addrs []string, g *cache.Group) {
	peers := cache.NewCacheServer(addr)
	peers.Set(addrs...)
	g.RegisterPeers(peers)
	log.Println("Cache is Running at", addr)
	log.Fatal(peers.Run())
}

// 在本机apiAddr上启动api网关服务
func StartAPIServer(apiAddr string, g *cache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := g.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
		},
	))
	log.Println("fontend server is running at :", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

var (
	port int
	api  bool
)

func init() {
	flag.IntVar(&port, "port", 8001, "the port of cache server")
	flag.BoolVar(&api, "api", false, "start a api server?")
}
func main() {
	flag.Parse()
	apiAddr := "http://localhost:9999"
	//
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	var addrs []string
	for _, addr := range addrMap {
		addrs = append(addrs, addr)
	}
	//创建一个缓存组,名字叫"scores",[]addrMap内的三个服务器都属于该同名缓存组集群内
	//它们逻辑上属于同一个分布式系统
	Cache := CreateGroup()
	if api {
		go StartAPIServer(apiAddr, Cache)
	}
	StartCacheServer(addrMap[port], addrs, Cache)
}
