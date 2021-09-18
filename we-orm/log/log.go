package log

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
)

//批量定义一组变量
var (
	//log.Lshortfile 支持显示文件名和方法
	errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m", log.LstdFlags|log.Lshortfile) //红色字体
	infoLog  = log.New(os.Stdout, "\033[34m[info]\033[0m", log.LstdFlags|log.Lshortfile)  //蓝色字体
	loggers  = []*log.Logger{errorLog, infoLog}
	mu       sync.Mutex //互斥锁
)

//向外暴露的四个方法
var (
	Error  = errorLog.Println
	Errorf = errorLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
)

//批量声明多个常量
const (
	//iota: 常量计数器
	infoLevel  = iota //0
	ErrorLevel        //1
	Disabled          //2
)

//控制log的等级
func SetLevel(level int) {
	mu.Lock()         //加锁
	defer mu.Unlock() //保证最终锁被释放

	for _, logger := range loggers {
		logger.SetOutput(os.Stdout)
	}

	if ErrorLevel < level {
		errorLog.SetOutput(ioutil.Discard)
	}
	if infoLevel < level {
		//若小于所设level,则将输出重定向到 io.Discard
		//即丢弃该输出
		infoLog.SetOutput(ioutil.Discard)
	}
}
