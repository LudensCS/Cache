package cache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/cache/"

// <BasePath>/<name>/<key>
type HTTPPool struct {
	//self example:http://127.0.0.1:8888
	Self     string
	BasePath string
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		Self:     self,
		BasePath: defaultBasePath,
	}
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
