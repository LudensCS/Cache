package main

import (
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

func main() {
	cache.NewGroup("scores", 2<<10, cache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		},
	))
	addr := "localhost:9999"
	peers := cache.NewHTTPPool(addr)
	log.Println("cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
