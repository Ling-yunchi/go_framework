package log

import (
	"log"
	"os"
	"sync"
)

var (
	//log.Lshortfile 支持显示文件名和方法
	errorLog = log.New(os.Stdout, "\033[31m[error]\033[0m", log.LstdFlags|log.Lshortfile) //红色字体
	infoLog  = log.New(os.Stdout, "\033[34m[error]\033[0m", log.LstdFlags|log.Lshortfile) //蓝色字体
	loggers  = []*log.Logger{errorLog, infoLog}
	mu       sync.Mutex
)

var (
	//向外暴露的四个方法
	Error  = errorLog.Println
	Errorf = errorLog.Printf
	Info   = infoLog.Println
	Infof  = infoLog.Printf
)
