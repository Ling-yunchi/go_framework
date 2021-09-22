package wecache

import (
	"fmt"
	"log"
	"sync"
)

//函数类型实现某一个接口，称之为接口型函数
//	方便使用者在调用时既能够传入函数作为参数
//	也能够传入实现了该接口的结构体作为参数

//Getter 缓存未命中时获取源数据的回调
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

//Group 是一个缓存命名空间,加载相关的数据
type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	mu     sync.RWMutex //读写锁
	groups = make(map[string]*Group)
)

//NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

//GetGroup returns the named group previously created with NewGroup,
//	or nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

//Get 查找缓存
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[WeCache] hit")
		return v, nil
	}
	return g.load(key)
}

//load 未找到缓存时加载缓存
//	单机场景下调用getLocally(key)
//	分布式场景下调用getFromPeer(key)
func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

//getLocally 本地获取数据,通过调用用户回调函数获取数据源
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

//populateCache 将源数据加入缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
