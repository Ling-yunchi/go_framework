package singleflight

import "sync"

//call 代表正在进行中,或已经结束的请求
//	使用 sync.WaitGroup 锁避免重入
type call struct {
	//sync.WaitGroup 等待任务完成锁
	//	Add(n): 增加n个需要完成的任务数量
	//	Wait(): 阻塞,等待 需要完成的任务数量 为0后继续往下执行代码
	//	Done():	完成一个任务,需要完成的任务数减一,相当于Add(-1)
	wg  sync.WaitGroup
	val interface{}
	err error
}

//Group 管理不同 key 的请求(call)
type Group struct {
	mu    sync.Mutex       //保护calls不被并发读写的锁
	calls map[string]*call //存储所有正在进行或已结束的请求
}

//Do 对于相同的key,无论 Do() 被调用几次, fn() 都只会被调用一次
//	等待 fn() 调用结束后,返回返回值
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock() //锁住calls
	//若calls不存在则初始化
	if g.calls == nil {
		g.calls = make(map[string]*call)
	}
	//判断calls中是否已经存在key对应的请求
	if c, ok := g.calls[key]; ok {
		g.mu.Unlock()       //存在则解锁calls,因为不需要对calls进行读写
		c.wg.Wait()         //等待该key对应call执行完毕
		return c.val, c.err //返回该key对应call的结果
	}

	c := new(call)   //新建一个请求
	c.wg.Add(1)      //设置解锁需完成的任务数量
	g.calls[key] = c //在calls中存储call
	g.mu.Unlock()    //结束对calls的读写,解锁calls

	c.val, c.err = fn() //调用回调函数,即call所对应的操作
	c.wg.Done()         //call请求完毕,完成任务数加一

	g.mu.Lock()          //请求完毕后要删除请求,锁住calls
	delete(g.calls, key) //删除key对应请求
	g.mu.Unlock()        //解锁

	return c.val, c.err //返回请求结果
}
