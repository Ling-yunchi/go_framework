package wego

import (
	"fmt"
	"strings"
)

//pattern 路由匹配模式: 基本网址后的路径

type node struct {
	pattern  string  //待匹配路由模式
	part     string  //路由中的一部分
	children []*node //子节点
	isWild   bool    //是否为非精确匹配
}

func (n *node) String() string {
	return fmt.Sprintf("node{pattern=%s, part=%s, isWild=%t}", n.pattern, n.part, n.isWild)
}

func (n *node) travel(list *[]*node) {
	if n.pattern != "" {
		*list = append(*list, n)
	}
	for _, child := range n.children {
		child.travel(list)
	}
}

//matchChild 在n的孩子节点中查找第一个匹配的节点,用于插入
func (n *node) matchChild(part string) *node {
	//优先匹配可精确匹配的节点
	for _, child := range n.children {
		if child.part == part {
			return child
		}
	}
	//未找到可精确匹配的节点才查找模糊匹配的节点
	for _, child := range n.children {
		if child.isWild {
			return child
		}
	}
	return nil
}

//matchChildren 在n的孩子节点中查找所有匹配成功的节点,用于查询
func (n *node) matchChildren(part string) []*node {
	//保证精确匹配的节点优先遍历
	nodes := make([]*node, 0)
	wildNodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part {
			nodes = append(nodes, child)
		} else if child.isWild {
			wildNodes = append(wildNodes, child)
		}
	}
	nodes = append(nodes, wildNodes...)
	return nodes
}

//insert 插入节点
func (n *node) insert(pattern string, parts []string, height int) {
	//patten: 待匹配路由路径  parts: 分割后的pattern的各部分  height: 节点深度  n: 当前匹配到的节点
	if len(parts) == height {
		//一个路径中的每一个part存储一层节点中
		//若深度与part的个数相同时,则已经匹配到了底层
		//在最下面的叶子节点存储整个匹配路径
		if n.pattern != "" {
			//若n存储的pattern不为空,则说明该节点已匹配路由规则,此时路由规则产生冲突
			panic(fmt.Sprintf("Route Conflict: %s : %s", n.pattern, pattern))
		}
		n.pattern = pattern
		return
	}
	//part: 当前层需要匹配的部分
	part := parts[height]
	//查找是否有匹配的节点
	child := n.matchChild(part)
	if child == nil {
		//若没有,则创建一个节点,加入当前节点的子节点
		child = &node{
			part:   part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		n.children = append(n.children, child)
	}
	//递归继续向下插入
	child.insert(pattern, parts, height+1)
}

//search 查找节点
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		//匹配到了底层或是当前节点保存的part的前缀为*时
		//如果n的pattern保存了路由规则,则返回n,否则返回nil
		if n.pattern == "" {
			return nil
		}
		return n
	}
	//part: 当前层需要查找的部分
	part := parts[height]
	//查找所有可以匹配的节点,包括动态匹配的节点
	children := n.matchChildren(part)
	for _, child := range children {
		//对每个可以匹配的节点递归向下进行查找
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}
