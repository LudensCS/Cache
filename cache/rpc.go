package cache

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/LudensCS/Cache/cache/cachepb"
	"github.com/LudensCS/Cache/cache/consistenthash"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const defaultReplicas = 50

type (
	//rpc服务器
	CacheServer struct {
		cachepb.UnimplementedGroupCacheServer
		Self    string //Self example : http://localhost:8888
		mutex   sync.Mutex
		peers   *consistenthash.Map
		Getters map[string]*CacheClient
	}
	//rpc客户端
	CacheClient struct {
		BaseURL string //BaseURL example : http://localhost:8888
	}
)

// 构造函数
func NewCacheServer(addr string) *CacheServer {
	return &CacheServer{
		Self:    addr,
		mutex:   sync.Mutex{},
		peers:   nil,
		Getters: make(map[string]*CacheClient),
	}
}

// 日志
func (CS *CacheServer) Log(format string, args ...any) {
	log.Printf("Server [%s] : %s\n", CS.Self, fmt.Sprintf(format, args...))
}
func (CS *CacheServer) Get(ctx context.Context, Req *cachepb.Request) (*cachepb.Response, error) {
	group := GetGroup(Req.GetGroup())
	if group == nil {
		return &cachepb.Response{}, status.Error(codes.Internal, "group not found")
	}
	value, err := group.Get(Req.GetKey())
	if err != nil {
		return &cachepb.Response{}, err
	}
	return &cachepb.Response{Value: value.ByteSlice()}, nil
}

// 注册分布式系统中的节点
func (CS *CacheServer) Set(peers ...string) {
	CS.mutex.Lock()
	defer CS.mutex.Unlock()
	if CS.peers == nil {
		CS.peers = consistenthash.New(defaultReplicas, nil)
	}
	CS.peers.Add(peers...)
	for _, peer := range peers {
		CS.Getters[peer] = &CacheClient{BaseURL: peer}
	}
}

// 利用一致性哈希选择远端节点
func (CS *CacheServer) PickPeer(key string) (PeerGetter, bool) {
	CS.mutex.Lock()
	defer CS.mutex.Unlock()
	if peer := CS.peers.Get(key); peer != "" && peer != CS.Self {
		CS.Log("Pick peer %s", peer)
		return CS.Getters[peer], true
	}
	return nil, false

}

// 启动rpc服务
func (CS *CacheServer) Run() error {
	S := grpc.NewServer()
	cachepb.RegisterGroupCacheServer(S, CS)
	listener, err := net.Listen("tcp", CS.Self[7:])
	if err != nil {
		return err
	}
	defer listener.Close()
	return S.Serve(listener)
}

// 启动rpc客户端调用
func (CC *CacheClient) Get(Req *cachepb.Request) (*cachepb.Response, error) {
	conn, err := grpc.NewClient(CC.BaseURL[7:], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return &cachepb.Response{}, err
	}
	defer conn.Close()
	client := cachepb.NewGroupCacheClient(conn)
	Resp, err := client.Get(context.Background(), Req)
	if err != nil {
		return &cachepb.Response{}, err
	}
	return Resp, nil
}
