package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/LudensCS/Cache/cache"
	"github.com/LudensCS/Cache/database/mysql"
	"github.com/LudensCS/Cache/middlewares/bloomfilter"
	"github.com/joho/godotenv"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

var db *gorm.DB
var dsn string

// CreateGroup 创立缓存组
func CreateGroup() *cache.Group {
	return cache.NewGroup("scores", 2<<10, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			row, err := mysql.Select(db, key)
			if err != nil {
				return nil, err
			}
			return row[0].Value, nil
		},
	))
}

// StartCacheServer 启动缓存服务
func StartCacheServer(addr string, addrs []string, g *cache.Group) {
	peers := cache.NewCacheServer(addr)
	peers.Set(addrs...)
	g.RegisterPeers(peers)
	log.Println("Cache is Running at", addr)
	log.Fatal(peers.Run())
}

// StartAPIServer 在本机apiAddr上启动api网关服务
func StartAPIServer(apiAddr string, g *cache.Group) {
	Filter := LoadDB()
	//example : http://apiAddr/api?key=xxx
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			if !Filter.Query(key) {
				err := status.Errorf(codes.NotFound, "%v not exist", key)
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
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

// LoadDB 将数据库中的数据加载到布隆过滤器
func LoadDB() *bloomfilter.Bloomfilter {
	rows, err := mysql.Select(db, "*")
	if err != nil {
		log.Fatal(err)
	}
	Filter := bloomfilter.New(len(rows))
	for _, row := range rows {
		Filter.Add(row.Key)
	}
	return Filter
}

var (
	port     int
	api      bool
	loaddata bool
)

func init() {
	flag.IntVar(&port, "port", 8000, "the port of cache server")
	flag.BoolVar(&api, "api", false, "start a api server?")
	flag.BoolVar(&loaddata, "load", false, "initial database with pre-datas")
	if err := godotenv.Load("./variables.env"); err != nil {
		log.Fatal(err)
	}
	dsn = fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?loc=Local&parseTime=true&charset=utf8mb4",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"), os.Getenv("DB_NAME"),
	)
}
func main() {
	flag.Parse()
	//数据库初始化
	var err error
	db, err = mysql.Register(dsn)
	if err != nil {
		log.Fatal(err)
	}
	if loaddata {
		var datas = []*mysql.Data{
			{Key: "Jack", Value: []byte("Admin")},
			{Key: "Lucy", Value: []byte("User")},
			{Key: "David", Value: []byte("User")},
		}
		err = mysql.Init(db, datas)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("local data has loaded")
		return
	}
	apiAddr := "http://localhost:9999"
	//分布式系统中各节点地址
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
