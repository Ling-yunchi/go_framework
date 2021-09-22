package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

//一致性哈希算法

//Hash maps bytes to uint32
type Hash func(data []byte) uint32

//Map contains all hashed keys
type Map struct {
	hash     Hash           //使用的Hash算法
	replicas int            //虚拟节点倍数,用于在节点数目较小时避免key的数据倾斜问题
	keys     []int          //哈希环 已排序
	hashMap  map[int]string //虚拟节点与真实节点的映射表
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		//若Hash算法未指定,默认使用crc32.ChecksumIEEE算法
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

//Add 添加真实节点/机器
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		//对每一个真实节点 key,对应创建 m.replicas 个虚拟节点
		//虚拟节点的名称是: strconv.Itoa(i) + key,即通过添加编号的方式区分不同虚拟节点
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			//增加虚拟节点和真实节点的映射关系
			m.hashMap[hash] = key
		}
	}
	//环上的哈希值排序
	sort.Ints(m.keys)
}

//Get 获取离所给key最近的机器的key,由该机器提供所需数据
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	//二分查找对应的机器节点
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
