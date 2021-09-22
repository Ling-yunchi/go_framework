package lru

import "container/list"

type Cache struct {
	maxBytes  int64 //允许使用的最大内存,为0时不做限制
	nowBytes  int64 //当前已使用的内存
	list      *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value) //记录被移除时的回调函数，可以为 nil
}

//entry 双向链表节点的数据类型
type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, OnEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		list:      list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: OnEvicted,
	}
}

//LRU 最近最少使用，相对于仅考虑时间因素的 FIFO 和仅考虑访问频率的 LFU，LRU 算法可以认为是相对平衡的一种淘汰算法。
//	LRU 认为，如果数据最近被访问过，那么将来被访问的概率也会更高。
//	LRU 算法的实现非常简单，维护一个队列，如果某条记录被访问了，则移动到队尾，那么队首则是最近最少访问的数据，淘汰该条记录即可。

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		//如果键对应的链表节点存在，则将对应节点移动到队尾，并返回查找到的值
		c.list.MoveToFront(ele) //此处使用双向链表模拟队列,先进为队首,后进为队尾
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

//RemoveOldest 缓存淘汰
func (c *Cache) RemoveOldest() {
	//获取队首元素
	ele := c.list.Back()
	if ele != nil {
		c.list.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nowBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

//Add 新增/修改记录
//	如果键存在，则更新对应节点的值，并将该节点移到队尾。
//	不存在则是新增场景，首先队尾添加新节点 &entry{key, value}, 并字典中添加 key 和节点的映射关系。
//	更新 c.nowBytes，如果超过了设定的最大值 c.maxBytes，则移除最少访问的节点
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.list.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nowBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.list.PushFront(&entry{key: key, value: value})
		c.cache[key] = ele
		c.nowBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nowBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.list.Len()
}
