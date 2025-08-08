package cache

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/LudensCS/Cache/cache/consistenthash"
)

const (
	defaultBasePath = "/cache/"
	defaultReplicas = 50
)

// <BasePath>/<name>/<key>
type (
	HTTPPool struct {
		Self        string                 //self example:http://127.0.0.1:8888
		BasePath    string                 // default : "/cache/"
		mutex       sync.Mutex             //protect peers and httpGetters
		peers       *consistenthash.Map    //select peer
		httpGetters map[string]*HTTPGetter //key : URL
	}
	HTTPGetter struct {
		BaseURL string
	}
)

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		Self:        self,
		BasePath:    defaultBasePath,
		mutex:       sync.Mutex{},
		peers:       nil,
		httpGetters: make(map[string]*HTTPGetter),
	}
}

// update the pool's list of peers
func (pool *HTTPPool) Set(peers ...string) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	if pool.peers == nil {
		pool.peers = consistenthash.New(defaultReplicas, nil)
	}
	pool.peers.Add(peers...)
	pool.httpGetters = make(map[string]*HTTPGetter, len(peers))
	for _, peer := range peers {
		pool.httpGetters[peer] = &HTTPGetter{BaseURL: peer + pool.BasePath}
	}
}

// picks a peer according to key
func (pool *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	if peer := pool.peers.Get(key); peer != "" && peer != pool.Self {
		pool.Log("Pick peer %s", peer)
		return pool.httpGetters[peer], true
	}
	return nil, false
}

func (pool *HTTPPool) Log(format string, v ...any) {
	log.Printf("[Server %s] %s\n", pool.Self, fmt.Sprintf(format, v...))
}

func (pool *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, pool.BasePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	pool.Log("%s %s", r.Method, r.URL.Path)
	parts := strings.SplitN(r.URL.Path[len(pool.BasePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad Request", http.StatusBadRequest)
		return
	}
	groupName, key := parts[0], parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "group not found", http.StatusNotFound)
	}
	value, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(value.ByteSlice())
}
func (h *HTTPGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", h.BaseURL, url.QueryEscape(group), url.QueryEscape(key))
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", resp.StatusCode)
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body error:%v", err)
	}
	return bytes, nil
}
