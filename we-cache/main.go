package main

import (
	"fmt"
	"log"
	"net/http"
	"wecache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	wecache.NewGroup("scores", 2<<10, wecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:8080"
	peers := wecache.NewHTTPPool(addr)
	log.Println("wecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
